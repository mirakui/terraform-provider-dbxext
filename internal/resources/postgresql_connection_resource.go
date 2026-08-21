package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	dbclient "github.com/mirakui/terraform-provider-dbxext/internal/databricks"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*PostgreSQLConnectionResource)(nil)
var _ resource.ResourceWithConfigure = (*PostgreSQLConnectionResource)(nil)
var _ resource.ResourceWithImportState = (*PostgreSQLConnectionResource)(nil)

type PostgreSQLConnectionResource struct {
	client dbclient.ConnectionClient
}

// Attributes Terraform computes or leaves optional arrive as unknown or null,
// which plain Go types cannot represent, so they use framework types. Required
// attributes are always known in a plan or state and stay plain.
type PostgreSQLConnectionModel struct {
	ID                    types.String              `tfsdk:"id"`
	ConnectionID          types.String              `tfsdk:"connection_id"`
	FullName              types.String              `tfsdk:"full_name"`
	MetastoreID           types.String              `tfsdk:"metastore_id"`
	CredentialType        types.String              `tfsdk:"credential_type"`
	URL                   types.String              `tfsdk:"url"`
	CreatedAt             types.Int64               `tfsdk:"created_at"`
	CreatedBy             types.String              `tfsdk:"created_by"`
	UpdatedAt             types.Int64               `tfsdk:"updated_at"`
	UpdatedBy             types.String              `tfsdk:"updated_by"`
	ProvisioningInfo      types.Object              `tfsdk:"provisioning_info"`
	Name                  string                    `tfsdk:"name"`
	Host                  string                    `tfsdk:"host"`
	Port                  int64                     `tfsdk:"port"`
	User                  string                    `tfsdk:"user"`
	PasswordSecret        *PasswordSecretModel      `tfsdk:"password_secret"`
	PasswordSecretVersion int64                     `tfsdk:"password_secret_version"`
	Comment               types.String              `tfsdk:"comment"`
	ReadOnly              *bool                     `tfsdk:"read_only"`
	Owner                 types.String              `tfsdk:"owner"`
	Properties            map[string]string         `tfsdk:"properties"`
	EnvironmentSettings   *EnvironmentSettingsModel `tfsdk:"environment_settings"`
}

var provisioningInfoAttributeTypes = map[string]attr.Type{
	"state": types.StringType,
}

type PasswordSecretModel struct {
	Scope string `tfsdk:"scope"`
	Key   string `tfsdk:"key"`
}

type EnvironmentSettingsModel struct {
	EnvironmentVersion string   `tfsdk:"environment_version"`
	JavaDependencies   []string `tfsdk:"java_dependencies"`
}

type ProvisioningInfoModel struct {
	State string `tfsdk:"state"`
}

func NewPostgreSQLConnectionResource() resource.Resource {
	return &PostgreSQLConnectionResource{}
}

func (r *PostgreSQLConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_connection"
}

func (r *PostgreSQLConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages a Databricks PostgreSQL external connection using typed fields and a Databricks secret reference for the password.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform state identity for the Databricks connection.",
			},
			"connection_id": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Databricks connection identifier.",
			},
			"full_name": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Databricks full connection name.",
			},
			"metastore_id": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Databricks metastore identifier.",
			},
			"credential_type": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Databricks credential type.",
			},
			"url": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Databricks-derived remote data source URL.",
			},
			"created_at": rschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp in epoch milliseconds.",
			},
			"created_by": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creator principal.",
			},
			"updated_at": rschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Last update timestamp in epoch milliseconds.",
			},
			"updated_by": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last updater principal.",
			},
			"provisioning_info": rschema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Databricks provisioning status.",
				Attributes: map[string]rschema.Attribute{
					"state": rschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Databricks provisioning state.",
					},
				},
			},
			"name": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Databricks connection name.",
			},
			"host": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "PostgreSQL host.",
			},
			"port": rschema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "PostgreSQL port from 1 through 65535.",
			},
			"user": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "PostgreSQL username stored as non-secret metadata.",
			},
			"password_secret_version": rschema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Positive version marker used to reapply the Databricks secret reference.",
			},
			"comment": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Databricks connection comment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"read_only": rschema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Databricks read-only connection setting.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"owner": rschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Databricks connection owner.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"properties": rschema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Non-secret Databricks connection properties.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]rschema.Block{
			"environment_settings": rschema.SingleNestedBlock{
				MarkdownDescription: "Databricks connection environment settings.",
				Attributes: map[string]rschema.Attribute{
					"environment_version": rschema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Databricks environment version.",
					},
					"java_dependencies": rschema.ListAttribute{
						Optional:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Java dependency coordinates for the Databricks connection environment.",
					},
				},
			},
			"password_secret": rschema.SingleNestedBlock{
				MarkdownDescription: "Databricks secret reference for the PostgreSQL password.",
				Attributes: map[string]rschema.Attribute{
					"scope": rschema.StringAttribute{
						Required:            true,
						Sensitive:           true,
						MarkdownDescription: "Databricks secret scope.",
					},
					"key": rschema.StringAttribute{
						Required:            true,
						Sensitive:           true,
						MarkdownDescription: "Databricks secret key.",
					},
				},
			},
		},
	}
}

func (r *PostgreSQLConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing Databricks client", "The provider did not configure a Databricks connection client.")
		return
	}

	var plan PostgreSQLConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := CreatePostgreSQLConnection(ctx, r.client, plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create PostgreSQL connection", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PostgreSQLConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing Databricks client", "The provider did not configure a Databricks connection client.")
		return
	}

	var state PostgreSQLConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, exists, err := ReadPostgreSQLConnection(ctx, r.client, state)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read PostgreSQL connection", err.Error())
		return
	}
	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PostgreSQLConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing Databricks client", "The provider did not configure a Databricks connection client.")
		return
	}

	var prior PostgreSQLConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan PostgreSQLConnectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := UpdatePostgreSQLConnection(ctx, r.client, prior, plan)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update PostgreSQL connection", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PostgreSQLConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Missing Databricks client", "The provider did not configure a Databricks connection client.")
		return
	}

	var state PostgreSQLConnectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := DeletePostgreSQLConnection(ctx, r.client, state.Name); err != nil {
		resp.Diagnostics.AddError("Unable to delete PostgreSQL connection", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *PostgreSQLConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	switch client := req.ProviderData.(type) {
	case dbclient.Client:
		r.client = client.Connections()
	case dbclient.ConnectionClient:
		r.client = client
	default:
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected Databricks client, got %T.", req.ProviderData))
	}
}

func (r *PostgreSQLConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func ValidatePostgreSQLConnectionModel(ctx context.Context, model PostgreSQLConnectionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	addRequiredStringDiagnostic := func(name, value string) {
		if strings.TrimSpace(value) == "" {
			diags.AddError("Invalid PostgreSQL connection configuration", fmt.Sprintf("%s must be non-empty.", name))
		}
	}

	addRequiredStringDiagnostic("name", model.Name)
	addRequiredStringDiagnostic("host", model.Host)
	addRequiredStringDiagnostic("user", model.User)
	if model.PasswordSecret == nil {
		diags.AddError("Invalid PostgreSQL connection configuration", "password_secret must be set.")
	} else {
		addRequiredStringDiagnostic("password_secret.scope", model.PasswordSecret.Scope)
		addRequiredStringDiagnostic("password_secret.key", model.PasswordSecret.Key)
	}

	if model.Port < 1 || model.Port > 65535 {
		diags.AddError("Invalid PostgreSQL connection configuration", "port must be between 1 and 65535.")
	}
	if model.PasswordSecretVersion < 1 {
		diags.AddError("Invalid PostgreSQL connection configuration", "password_secret_version must be a positive integer.")
	}

	return diags
}

func CreatePostgreSQLConnection(ctx context.Context, client dbclient.ConnectionClient, model PostgreSQLConnectionModel) (PostgreSQLConnectionModel, error) {
	if client == nil {
		return PostgreSQLConnectionModel{}, fmt.Errorf("Databricks connection client is required")
	}

	diags := ValidatePostgreSQLConnectionModel(ctx, model)
	if diags.HasError() {
		return PostgreSQLConnectionModel{}, fmt.Errorf("invalid PostgreSQL connection configuration")
	}

	req, err := dbclient.BuildCreateConnectionRequest(dbclient.PostgreSQLConnectionConfig{
		Name: model.Name,
		Host: model.Host,
		Port: model.Port,
		User: model.User,
		PasswordSecret: dbclient.PasswordSecretReference{
			Scope: model.PasswordSecret.Scope,
			Key:   model.PasswordSecret.Key,
		},
		PasswordSecretVersion: model.PasswordSecretVersion,
	})
	if err != nil {
		return PostgreSQLConnectionModel{}, err
	}
	req.Comment = optionalString(model.Comment.ValueString())
	req.ReadOnly = model.ReadOnly
	req.Owner = strings.TrimSpace(model.Owner.ValueString())
	req.Properties = model.Properties

	remote, err := client.CreateConnection(ctx, req)
	if err != nil {
		return PostgreSQLConnectionModel{}, err
	}

	state := mergeConnectionInfo(model, remote)
	desiredOwner := strings.TrimSpace(model.Owner.ValueString())
	if desiredOwner != "" && state.Owner.ValueString() != desiredOwner {
		return UpdatePostgreSQLConnection(ctx, client, state, model)
	}

	return state, nil
}

func DeletePostgreSQLConnection(ctx context.Context, client dbclient.ConnectionClient, name string) error {
	if client == nil {
		return fmt.Errorf("Databricks connection client is required")
	}
	return client.DeleteConnection(ctx, name)
}

func ReadPostgreSQLConnection(ctx context.Context, client dbclient.ConnectionClient, state PostgreSQLConnectionModel) (PostgreSQLConnectionModel, bool, error) {
	remote, err := client.GetConnection(ctx, state.Name)
	if err != nil {
		if errors.Is(err, dbclient.ErrNotFound) {
			return PostgreSQLConnectionModel{}, false, nil
		}
		return PostgreSQLConnectionModel{}, true, err
	}
	return mergeConnectionInfo(state, remote), true, nil
}

func UpdatePostgreSQLConnection(ctx context.Context, client dbclient.ConnectionClient, prior PostgreSQLConnectionModel, plan PostgreSQLConnectionModel) (PostgreSQLConnectionModel, error) {
	if client == nil {
		return PostgreSQLConnectionModel{}, fmt.Errorf("Databricks connection client is required")
	}

	if diags := ValidatePostImportUpdateReady(ctx, plan); diags.HasError() {
		return PostgreSQLConnectionModel{}, fmt.Errorf("PostgreSQL connection update requires user and password secret metadata")
	}

	req, err := dbclient.BuildUpdateConnectionRequest(prior.Name, dbclient.PostgreSQLConnectionConfig{
		Name: plan.Name,
		Host: plan.Host,
		Port: plan.Port,
		User: plan.User,
		PasswordSecret: dbclient.PasswordSecretReference{
			Scope: plan.PasswordSecret.Scope,
			Key:   plan.PasswordSecret.Key,
		},
		PasswordSecretVersion: plan.PasswordSecretVersion,
		Owner:                 plan.Owner.ValueString(),
		EnvironmentSettings:   toDBClientEnvironmentSettings(plan.EnvironmentSettings),
	})
	if err != nil {
		return PostgreSQLConnectionModel{}, err
	}

	remote, err := client.UpdateConnection(ctx, prior.Name, req)
	if err != nil {
		return PostgreSQLConnectionModel{}, err
	}

	return mergeConnectionInfo(plan, remote), nil
}

func ValidatePostImportUpdateReady(ctx context.Context, model PostgreSQLConnectionModel) diag.Diagnostics {
	return ValidatePostgreSQLConnectionModel(ctx, model)
}

func PostgreSQLConnectionFieldRequiresReplacement(field string) bool {
	switch field {
	case "comment", "properties", "read_only":
		return true
	default:
		return false
	}
}

func PostgreSQLConnectionPasswordSecretVersionChanged(prior PostgreSQLConnectionModel, plan PostgreSQLConnectionModel) bool {
	return prior.PasswordSecretVersion != plan.PasswordSecretVersion
}

func mergeConnectionInfo(model PostgreSQLConnectionModel, remote dbclient.ConnectionInfo) PostgreSQLConnectionModel {
	if remote.Name != "" {
		model.ID = types.StringValue(remote.Name)
		model.Name = remote.Name
	}
	if model.ID.ValueString() == "" {
		model.ID = types.StringValue(model.Name)
	}
	model.ConnectionID = types.StringValue(remote.ID)
	model.FullName = types.StringValue(remote.FullName)
	model.MetastoreID = types.StringValue(remote.MetastoreID)
	model.CredentialType = types.StringValue(remote.CredentialType)
	model.URL = types.StringValue(remote.URL)
	model.CreatedAt = types.Int64Value(remote.CreatedAt)
	model.CreatedBy = types.StringValue(remote.CreatedBy)
	model.UpdatedAt = types.Int64Value(remote.UpdatedAt)
	model.UpdatedBy = types.StringValue(remote.UpdatedBy)
	if remote.ProvisioningInfo != nil {
		model.ProvisioningInfo = types.ObjectValueMust(provisioningInfoAttributeTypes, map[string]attr.Value{
			"state": types.StringValue(remote.ProvisioningInfo.State),
		})
	} else {
		// Computed attributes must not stay unknown once apply finishes.
		model.ProvisioningInfo = types.ObjectNull(provisioningInfoAttributeTypes)
	}
	if remote.Comment != "" {
		model.Comment = types.StringValue(remote.Comment)
	} else if model.Comment.IsUnknown() {
		model.Comment = types.StringNull()
	}
	if remote.ReadOnly != nil {
		model.ReadOnly = remote.ReadOnly
	}

	if remote.Options != nil {
		if host := remote.Options["host"]; host != "" {
			model.Host = host
		}
		if port := remote.Options["port"]; port != "" {
			if parsed, err := strconv.ParseInt(port, 10, 64); err == nil {
				model.Port = parsed
			}
		}
		if user := remote.Options["user"]; user != "" {
			model.User = user
		}
	}
	if remote.Owner != "" {
		model.Owner = types.StringValue(remote.Owner)
	} else if model.Owner.IsUnknown() {
		model.Owner = types.StringNull()
	}
	if remote.Properties != nil {
		model.Properties = remote.Properties
	}
	if remote.EnvironmentSettings != nil {
		model.EnvironmentSettings = &EnvironmentSettingsModel{
			EnvironmentVersion: remote.EnvironmentSettings.EnvironmentVersion,
			JavaDependencies:   remote.EnvironmentSettings.JavaDependencies,
		}
	}

	return model
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	return &trimmed
}

func toDBClientEnvironmentSettings(settings *EnvironmentSettingsModel) *dbclient.EnvironmentSettings {
	if settings == nil {
		return nil
	}
	return &dbclient.EnvironmentSettings{
		EnvironmentVersion: strings.TrimSpace(settings.EnvironmentVersion),
		JavaDependencies:   settings.JavaDependencies,
	}
}
