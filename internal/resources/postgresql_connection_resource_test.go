package resources

import (
	"context"
	"fmt"
	"strings"
	"testing"

	dbclient "github.com/mirakui/terraform-provider-dbxext/internal/databricks"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestPostgreSQLConnectionSchemaRequiresTypedFields(t *testing.T) {
	t.Parallel()

	schema := postgreSQLConnectionSchema(t)

	for _, name := range []string{"name", "host", "user"} {
		attr, ok := schema.Attributes[name]
		if !ok {
			t.Fatalf("expected schema attribute %q", name)
		}
		stringAttr, ok := attr.(rschema.StringAttribute)
		if !ok {
			t.Fatalf("expected %q to be a string attribute, got %T", name, attr)
		}
		if !stringAttr.Required {
			t.Fatalf("expected %q to be required", name)
		}
	}

	for _, name := range []string{"port", "password_secret_version"} {
		attr, ok := schema.Attributes[name]
		if !ok {
			t.Fatalf("expected schema attribute %q", name)
		}
		intAttr, ok := attr.(rschema.Int64Attribute)
		if !ok {
			t.Fatalf("expected %q to be an int64 attribute, got %T", name, attr)
		}
		if !intAttr.Required {
			t.Fatalf("expected %q to be required", name)
		}
	}

	block, ok := schema.Blocks["password_secret"]
	if !ok {
		t.Fatal("expected password_secret block")
	}
	secretBlock, ok := block.(rschema.SingleNestedBlock)
	if !ok {
		t.Fatalf("expected password_secret to be a single nested block, got %T", block)
	}
	for _, name := range []string{"scope", "key"} {
		attr, ok := secretBlock.Attributes[name]
		if !ok {
			t.Fatalf("expected password_secret.%s attribute", name)
		}
		stringAttr, ok := attr.(rschema.StringAttribute)
		if !ok {
			t.Fatalf("expected password_secret.%s to be a string attribute, got %T", name, attr)
		}
		if !stringAttr.Required {
			t.Fatalf("expected password_secret.%s to be required", name)
		}
	}
}

func TestPostgreSQLConnectionSchemaDoesNotExposeGenericConnectionFields(t *testing.T) {
	t.Parallel()

	schema := postgreSQLConnectionSchema(t)

	for _, name := range []string{"connection_type", "options", "password"} {
		if _, ok := schema.Attributes[name]; ok {
			t.Fatalf("unsupported attribute %q must not be exposed", name)
		}
		if _, ok := schema.Blocks[name]; ok {
			t.Fatalf("unsupported block %q must not be exposed", name)
		}
	}
}

func TestPostgreSQLConnectionSchemaAvoidsOwnerDriftAndSecretDisclosure(t *testing.T) {
	t.Parallel()

	schema := postgreSQLConnectionSchema(t)

	ownerAttr, ok := schema.Attributes["owner"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("expected owner to be a string attribute, got %T", schema.Attributes["owner"])
	}
	if !ownerAttr.Optional || !ownerAttr.Computed {
		t.Fatalf("expected owner to be optional and computed, got optional=%t computed=%t", ownerAttr.Optional, ownerAttr.Computed)
	}
	if len(ownerAttr.PlanModifiers) == 0 {
		t.Fatal("expected owner to preserve prior state when Terraform cannot know the Databricks default owner")
	}

	secretBlock := schema.Blocks["password_secret"].(rschema.SingleNestedBlock)
	for _, name := range []string{"scope", "key"} {
		attr := secretBlock.Attributes[name].(rschema.StringAttribute)
		if !attr.Sensitive {
			t.Fatalf("expected password_secret.%s to be sensitive metadata", name)
		}
	}
}

func TestPostgreSQLConnectionSchemaDoesNotExposeUnsupportedProviderConfig(t *testing.T) {
	t.Parallel()

	schema := postgreSQLConnectionSchema(t)

	if _, ok := schema.Blocks["provider_config"]; ok {
		t.Fatal("provider_config must not be exposed until the Databricks connection API supports it")
	}
	if PostgreSQLConnectionFieldRequiresReplacement("provider_config") {
		t.Fatal("provider_config must not be part of lifecycle rules when it is not exposed")
	}
}

func TestPostgreSQLConnectionValidationRejectsInvalidRequiredValues(t *testing.T) {
	t.Parallel()

	valid := PostgreSQLConnectionModel{
		Name:                  "psql",
		Host:                  "postgres.example.com",
		Port:                  5432,
		User:                  "postgres_user",
		PasswordSecret:        PasswordSecretModel{Scope: "scope", Key: "password"},
		PasswordSecretVersion: 1,
	}

	cases := map[string]PostgreSQLConnectionModel{
		"blank name":              withString(valid, "name", " "),
		"blank host":              withString(valid, "host", " "),
		"blank user":              withString(valid, "user", " "),
		"blank secret scope":      withString(valid, "scope", " "),
		"blank secret key":        withString(valid, "key", " "),
		"zero port":               withInt(valid, "port", 0),
		"port above maximum":      withInt(valid, "port", 65536),
		"zero secret version":     withInt(valid, "password_secret_version", 0),
		"negative secret version": withInt(valid, "password_secret_version", -1),
	}

	for name, model := range cases {
		t.Run(name, func(t *testing.T) {
			diags := ValidatePostgreSQLConnectionModel(context.Background(), model)
			if !diags.HasError() {
				t.Fatalf("expected validation diagnostics for %s", name)
			}
		})
	}

	diags := ValidatePostgreSQLConnectionModel(context.Background(), valid)
	if diags.HasError() {
		t.Fatalf("expected valid model, got diagnostics: %s", diags.Errors())
	}
}

func TestPostgreSQLConnectionCreateUsesSecretBackedTypedOptions(t *testing.T) {
	t.Parallel()

	rawPassword := strings.Join([]string{"dbxext", "raw", "password", "sentinel"}, "-")
	client := &mockConnectionClient{}

	state, err := CreatePostgreSQLConnection(context.Background(), client, PostgreSQLConnectionModel{
		Name:                  "psql",
		Host:                  "postgres.example.com",
		Port:                  5432,
		User:                  "postgres_user",
		PasswordSecret:        PasswordSecretModel{Scope: "scope", Key: "password"},
		PasswordSecretVersion: 1,
	})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if client.created.ConnectionType != "POSTGRESQL" {
		t.Fatalf("expected POSTGRESQL create request, got %q", client.created.ConnectionType)
	}
	if client.created.Options["password"] != "secret('scope', 'password')" {
		t.Fatalf("expected password secret expression, got %q", client.created.Options["password"])
	}
	if strings.Contains(fmt.Sprintf("%#v %#v", client.created, state), rawPassword) {
		t.Fatal("raw password sentinel leaked into request or state")
	}
	if state.ID != "psql" || state.ConnectionID != "connection-id" {
		t.Fatalf("unexpected created state: %#v", state)
	}
}

func TestPostgreSQLConnectionCreateAppliesConfiguredOwnerAfterCreate(t *testing.T) {
	t.Parallel()

	client := &mockConnectionClient{
		createOwner: "creator@example.com",
		updateOwner: "data-owner@example.com",
	}
	model := validPostgreSQLConnectionModel()
	model.Owner = "data-owner@example.com"

	state, err := CreatePostgreSQLConnection(context.Background(), client, model)
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if client.updatedName != "psql" {
		t.Fatalf("expected create to update owner after remote creation, got update name %q", client.updatedName)
	}
	if client.updated.Owner != "data-owner@example.com" {
		t.Fatalf("expected owner update request, got %q", client.updated.Owner)
	}
	if client.updated.Options["password"] != "secret('scope', 'password')" {
		t.Fatalf("expected owner update to preserve password secret reference, got %q", client.updated.Options["password"])
	}
	if state.Owner != "data-owner@example.com" {
		t.Fatalf("expected final state owner data-owner@example.com, got %q", state.Owner)
	}
}

func TestPostgreSQLConnectionDeleteUsesConnectionName(t *testing.T) {
	t.Parallel()

	client := &mockConnectionClient{}

	if err := DeletePostgreSQLConnection(context.Background(), client, "psql"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if client.deletedName != "psql" {
		t.Fatalf("expected delete name psql, got %q", client.deletedName)
	}
}

func TestPostgreSQLConnectionLifecycleReplacementDecisions(t *testing.T) {
	t.Parallel()

	replacementFields := []string{"comment", "properties", "read_only"}
	for _, field := range replacementFields {
		if !PostgreSQLConnectionFieldRequiresReplacement(field) {
			t.Fatalf("expected %s to require replacement", field)
		}
	}

	inPlaceFields := []string{
		"name",
		"host",
		"port",
		"user",
		"owner",
		"environment_settings",
		"password_secret",
		"password_secret_version",
	}
	for _, field := range inPlaceFields {
		if PostgreSQLConnectionFieldRequiresReplacement(field) {
			t.Fatalf("expected %s to update in place", field)
		}
	}
}

func TestPostgreSQLConnectionUpdatePreservesPasswordSecretReference(t *testing.T) {
	t.Parallel()

	client := &mockConnectionClient{}
	prior := validPostgreSQLConnectionModel()
	plan := prior
	plan.Host = "postgres-new.example.com"
	plan.Port = 5433
	plan.User = "postgres_user_new"
	plan.Owner = "data-owner@example.com"
	plan.EnvironmentSettings = &EnvironmentSettingsModel{
		EnvironmentVersion: "14.2",
		JavaDependencies:   []string{"org.postgresql:postgresql:42.7.4"},
	}

	state, err := UpdatePostgreSQLConnection(context.Background(), client, prior, plan)
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if client.updatedName != "psql" {
		t.Fatalf("expected update name psql, got %q", client.updatedName)
	}
	if client.updated.Options["password"] != "secret('scope', 'password')" {
		t.Fatalf("expected update to preserve password secret reference, got %q", client.updated.Options["password"])
	}
	if client.updated.Options["host"] != "postgres-new.example.com" {
		t.Fatalf("expected updated host option, got %q", client.updated.Options["host"])
	}
	if state.Host != "postgres-new.example.com" || state.Port != 5433 || state.User != "postgres_user_new" {
		t.Fatalf("unexpected updated state: %#v", state)
	}
}

func TestPostgreSQLConnectionImportRequiresSecretMetadataBeforeManagedUpdate(t *testing.T) {
	t.Parallel()

	imported := PostgreSQLConnectionModel{
		ID:           "psql",
		ConnectionID: "connection-id",
		Name:         "psql",
		Host:         "postgres.example.com",
		Port:         5432,
	}
	if diags := ValidatePostImportUpdateReady(context.Background(), imported); !diags.HasError() {
		t.Fatal("expected imported state without user/password metadata to be rejected before managed update")
	}

	complete := imported
	complete.User = "postgres_user"
	complete.PasswordSecret = PasswordSecretModel{Scope: "scope", Key: "password"}
	complete.PasswordSecretVersion = 1
	if diags := ValidatePostImportUpdateReady(context.Background(), complete); diags.HasError() {
		t.Fatalf("expected completed imported state to be update-ready, got diagnostics: %s", diags.Errors())
	}
}

func TestPostgreSQLConnectionPasswordSecretVersionChangeTriggersPasswordRefresh(t *testing.T) {
	t.Parallel()

	prior := validPostgreSQLConnectionModel()
	plan := prior
	plan.PasswordSecretVersion = 2

	if !PostgreSQLConnectionPasswordSecretVersionChanged(prior, plan) {
		t.Fatal("expected password_secret_version change to be detected")
	}
	if PostgreSQLConnectionFieldRequiresReplacement("password_secret_version") {
		t.Fatal("password_secret_version must update in place")
	}

	client := &mockConnectionClient{}
	state, err := UpdatePostgreSQLConnection(context.Background(), client, prior, plan)
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if client.updated.Options["password"] != "secret('scope', 'password')" {
		t.Fatalf("expected rotation update to reapply password secret reference, got %q", client.updated.Options["password"])
	}
	if state.PasswordSecretVersion != 2 {
		t.Fatalf("expected state version 2, got %d", state.PasswordSecretVersion)
	}
}

func TestPostgreSQLConnectionReadMapsComputedRemoteFields(t *testing.T) {
	t.Parallel()

	readOnly := true
	state := mergeConnectionInfo(validPostgreSQLConnectionModel(), dbclient.ConnectionInfo{
		ID:             "connection-id",
		Name:           "psql",
		FullName:       "metastore.psql",
		MetastoreID:    "metastore-id",
		CredentialType: "USERNAME_PASSWORD",
		URL:            "postgresql://postgres.example.com:5432",
		CreatedAt:      1000,
		CreatedBy:      "creator@example.com",
		UpdatedAt:      2000,
		UpdatedBy:      "updater@example.com",
		Comment:        "remote comment",
		ReadOnly:       &readOnly,
		Properties:     map[string]string{"purpose": "analytics"},
		EnvironmentSettings: &dbclient.EnvironmentSettings{
			EnvironmentVersion: "14.2",
			JavaDependencies:   []string{"org.postgresql:postgresql:42.7.4"},
		},
		ProvisioningInfo: &dbclient.ProvisioningInfo{State: "ACTIVE"},
	})

	if state.FullName != "metastore.psql" ||
		state.MetastoreID != "metastore-id" ||
		state.CredentialType != "USERNAME_PASSWORD" ||
		state.URL != "postgresql://postgres.example.com:5432" ||
		state.CreatedAt != 1000 ||
		state.CreatedBy != "creator@example.com" ||
		state.UpdatedAt != 2000 ||
		state.UpdatedBy != "updater@example.com" {
		t.Fatalf("computed fields were not mapped from remote state: %#v", state)
	}
	if state.Comment != "remote comment" || state.ReadOnly == nil || !*state.ReadOnly {
		t.Fatalf("optional metadata was not mapped from remote state: %#v", state)
	}
	if state.Properties["purpose"] != "analytics" {
		t.Fatalf("properties were not mapped from remote state: %#v", state.Properties)
	}
	if state.EnvironmentSettings == nil || state.EnvironmentSettings.EnvironmentVersion != "14.2" {
		t.Fatalf("environment settings were not mapped from remote state: %#v", state.EnvironmentSettings)
	}
	if state.ProvisioningInfo == nil || state.ProvisioningInfo.State != "ACTIVE" {
		t.Fatalf("provisioning info was not mapped from remote state: %#v", state.ProvisioningInfo)
	}
}

func TestPostgreSQLConnectionReadTreatsRemoteNotFoundAsGone(t *testing.T) {
	t.Parallel()

	client := &mockConnectionClient{getErr: dbclient.ErrNotFound}

	_, exists, err := ReadPostgreSQLConnection(context.Background(), client, validPostgreSQLConnectionModel())
	if err != nil {
		t.Fatalf("expected not found to remove state without an error, got %v", err)
	}
	if exists {
		t.Fatal("expected remote not found to mark the resource absent")
	}
}

func postgreSQLConnectionSchema(t *testing.T) rschema.Schema {
	t.Helper()

	resourceUnderTest := NewPostgreSQLConnectionResource()
	var resp resource.SchemaResponse
	resourceUnderTest.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %s", resp.Diagnostics.Errors())
	}

	return resp.Schema
}

func validPostgreSQLConnectionModel() PostgreSQLConnectionModel {
	return PostgreSQLConnectionModel{
		ID:                    "psql",
		ConnectionID:          "connection-id",
		Name:                  "psql",
		Host:                  "postgres.example.com",
		Port:                  5432,
		User:                  "postgres_user",
		PasswordSecret:        PasswordSecretModel{Scope: "scope", Key: "password"},
		PasswordSecretVersion: 1,
	}
}

func withString(model PostgreSQLConnectionModel, field string, value string) PostgreSQLConnectionModel {
	switch field {
	case "name":
		model.Name = value
	case "host":
		model.Host = value
	case "user":
		model.User = value
	case "scope":
		model.PasswordSecret.Scope = value
	case "key":
		model.PasswordSecret.Key = value
	}
	return model
}

func withInt(model PostgreSQLConnectionModel, field string, value int64) PostgreSQLConnectionModel {
	switch field {
	case "port":
		model.Port = value
	case "password_secret_version":
		model.PasswordSecretVersion = value
	}
	return model
}

type mockConnectionClient struct {
	created     dbclient.ConnectionRequest
	updated     dbclient.ConnectionRequest
	updatedName string
	deletedName string
	createOwner string
	updateOwner string
	getErr      error
}

func (m *mockConnectionClient) CreateConnection(ctx context.Context, req dbclient.ConnectionRequest) (dbclient.ConnectionInfo, error) {
	m.created = req
	return dbclient.ConnectionInfo{
		ID:             "connection-id",
		Name:           req.Name,
		ConnectionType: req.ConnectionType,
		Options:        req.Options,
		Owner:          m.createOwner,
	}, nil
}

func (m *mockConnectionClient) GetConnection(ctx context.Context, name string) (dbclient.ConnectionInfo, error) {
	if m.getErr != nil {
		return dbclient.ConnectionInfo{}, m.getErr
	}
	return dbclient.ConnectionInfo{
		ID:   "connection-id",
		Name: name,
	}, nil
}

func (m *mockConnectionClient) UpdateConnection(ctx context.Context, name string, req dbclient.ConnectionRequest) (dbclient.ConnectionInfo, error) {
	m.updatedName = name
	m.updated = req
	owner := req.Owner
	if m.updateOwner != "" {
		owner = m.updateOwner
	}
	return dbclient.ConnectionInfo{
		ID:             "connection-id",
		Name:           req.Name,
		ConnectionType: req.ConnectionType,
		Options:        req.Options,
		Owner:          owner,
	}, nil
}

func (m *mockConnectionClient) DeleteConnection(ctx context.Context, name string) error {
	m.deletedName = name
	return nil
}
