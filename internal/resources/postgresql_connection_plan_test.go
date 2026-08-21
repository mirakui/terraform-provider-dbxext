package resources

import (
	"context"
	"testing"

	dbclient "github.com/mirakui/terraform-provider-dbxext/internal/databricks"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// createPlanValue builds the object value Terraform hands to Create: every
// configured attribute is known, every computed attribute is still unknown.
func createPlanValue(t *testing.T, objectType tftypes.Object) tftypes.Value {
	t.Helper()

	known := map[string]tftypes.Value{
		"name":                    tftypes.NewValue(tftypes.String, "dbxext_plan_conn"),
		"host":                    tftypes.NewValue(tftypes.String, "db.example.com"),
		"port":                    tftypes.NewValue(tftypes.Number, 5432),
		"user":                    tftypes.NewValue(tftypes.String, "postgres_user"),
		"password_secret_version": tftypes.NewValue(tftypes.Number, 1),
		"password_secret": tftypes.NewValue(
			objectType.AttributeTypes["password_secret"],
			map[string]tftypes.Value{
				"scope": tftypes.NewValue(tftypes.String, "database"),
				"key":   tftypes.NewValue(tftypes.String, "postgres-password"),
			},
		),
	}

	// Optional attributes the practitioner left out of the configuration.
	nullNames := map[string]struct{}{
		"comment":              {},
		"read_only":            {},
		"properties":           {},
		"environment_settings": {},
	}

	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attrType := range objectType.AttributeTypes {
		value, configured := known[name]
		switch {
		case configured:
			attributes[name] = value
		case hasKey(nullNames, name):
			attributes[name] = tftypes.NewValue(attrType, nil)
		default:
			attributes[name] = tftypes.NewValue(attrType, tftypes.UnknownValue)
		}
	}

	return tftypes.NewValue(objectType, attributes)
}

func hasKey(set map[string]struct{}, name string) bool {
	_, ok := set[name]
	return ok
}

func TestPostgreSQLConnectionModelReadsCreatePlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schema := postgreSQLConnectionSchema(t)

	objectType, ok := schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("expected schema to describe an object type, got %T", schema.Type().TerraformType(ctx))
	}

	plan := tfsdk.Plan{Raw: createPlanValue(t, objectType), Schema: schema}

	var model PostgreSQLConnectionModel
	if diags := plan.Get(ctx, &model); diags.HasError() {
		t.Fatalf("expected the create plan to decode into the model, got %v", diags)
	}

	if model.Name != "dbxext_plan_conn" {
		t.Fatalf("expected name to survive decoding, got %q", model.Name)
	}
	if model.PasswordSecret == nil || model.PasswordSecret.Scope != "database" {
		t.Fatalf("expected the password secret block to survive decoding, got %+v", model.PasswordSecret)
	}
	if !model.URL.IsUnknown() || !model.ProvisioningInfo.IsUnknown() {
		t.Fatalf("expected computed attributes to stay unknown, got url=%v provisioning_info=%v", model.URL, model.ProvisioningInfo)
	}
	if !model.Comment.IsNull() {
		t.Fatalf("expected an omitted optional attribute to stay null, got %v", model.Comment)
	}
}

// A remote connection that reports no owner, comment or provisioning info must
// still leave every computed attribute known, or Terraform rejects the apply.
func TestMergeConnectionInfoResolvesUnknownComputedAttributes(t *testing.T) {
	t.Parallel()

	plan := PostgreSQLConnectionModel{
		Name:                  "dbxext_plan_conn",
		Host:                  "db.example.com",
		Port:                  5432,
		User:                  "postgres_user",
		PasswordSecret:        &PasswordSecretModel{Scope: "database", Key: "postgres-password"},
		PasswordSecretVersion: 1,
		ID:                    types.StringUnknown(),
		ConnectionID:          types.StringUnknown(),
		FullName:              types.StringUnknown(),
		MetastoreID:           types.StringUnknown(),
		CredentialType:        types.StringUnknown(),
		URL:                   types.StringUnknown(),
		CreatedAt:             types.Int64Unknown(),
		CreatedBy:             types.StringUnknown(),
		UpdatedAt:             types.Int64Unknown(),
		UpdatedBy:             types.StringUnknown(),
		Owner:                 types.StringUnknown(),
		Comment:               types.StringUnknown(),
		ProvisioningInfo:      types.ObjectUnknown(provisioningInfoAttributeTypes),
	}

	state := mergeConnectionInfo(plan, dbclient.ConnectionInfo{Name: "dbxext_plan_conn"})

	for name, value := range map[string]attr.Value{
		"id":                state.ID,
		"connection_id":     state.ConnectionID,
		"full_name":         state.FullName,
		"metastore_id":      state.MetastoreID,
		"credential_type":   state.CredentialType,
		"url":               state.URL,
		"created_at":        state.CreatedAt,
		"created_by":        state.CreatedBy,
		"updated_at":        state.UpdatedAt,
		"updated_by":        state.UpdatedBy,
		"owner":             state.Owner,
		"comment":           state.Comment,
		"provisioning_info": state.ProvisioningInfo,
	} {
		if value.IsUnknown() {
			t.Fatalf("expected %s to be known after merging remote state", name)
		}
	}
}
