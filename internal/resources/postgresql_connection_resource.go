package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dbclient "github.com/mirakui/terraform-provider-dbxext/internal/databricks"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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

type PostgreSQLConnectionModel struct {
	ID                    string                    `tfsdk:"id"`
	ConnectionID          string                    `tfsdk:"connection_id"`
	FullName              string                    `tfsdk:"full_name"`
	MetastoreID           string                    `tfsdk:"metastore_id"`
	CredentialType        string                    `tfsdk:"credential_type"`
	URL                   string                    `tfsdk:"url"`
	CreatedAt             int64                     `tfsdk:"created_at"`
	CreatedBy             string                    `tfsdk:"created_by"`
	UpdatedAt             int64                     `tfsdk:"updated_at"`
	UpdatedBy             string                    `tfsdk:"updated_by"`
	ProvisioningInfo      *ProvisioningInfoModel    `tfsdk:"provisioning_info"`
	Name                  string                    `tfsdk:"name"`
	Host                  string                    `tfsdk:"host"`
	Port                  int64                     `tfsdk:"port"`
	User                  string                    `tfsdk:"user"`
	PasswordSecret        PasswordSecretModel       `tfsdk:"password_secret"`
	PasswordSecretVersion int64                     `tfsdk:"password_secret_version"`
	Comment               string                    `tfsdk:"comment"`
	ReadOnly              *bool                     `tfsdk:"read_only"`
	Owner                 string                    `tfsdk:"owner"`
	Properties            map[string]string         `tfsdk:"properties"`
	EnvironmentSettings   *EnvironmentSettingsModel `tfsdk:"environment_settings"`
	ProviderConfig        *ProviderConfigModel      `tfsdk:"provider_config"`
}

type PasswordSecretModel struct {
	Scope string `tfsdk:"scope"`
	Key   string `tfsdk:"key"`
}

type EnvironmentSettingsModel struct {
	EnvironmentVersion string   `tfsdk:"environment_version"`
	JavaDependencies   []string `tfsdk:"java_dependencies"`
}

type ProviderConfigModel struct {
	WorkspaceID int64 `tfsdk:"workspace_id"`
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
				MarkdownDescription: "Databricks connection owner.",
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
						MarkdownDescription: "Databricks secret scope.",
					},
					"key": rschema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Databricks secret key.",
					},
				},
			},
			"provider_config": rschema.SingleNestedBlock{
				MarkdownDescription: "Optional workspace routing metadata.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]rschema.Attribute{
					"workspace_id": rschema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "Databricks workspace ID.",
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

	remote, err := r.client.GetConnection(ctx, state.Name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read PostgreSQL connection", err.Error())
		return
	}

	state = mergeConnectionInfo(state, remote)
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
	addRequiredStringDiagnostic("password_secret.scope", model.PasswordSecret.Scope)
	addRequiredStringDiagnostic("password_secret.key", model.PasswordSecret.Key)

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
	req.Comment = optionalString(model.Comment)
	req.ReadOnly = model.ReadOnly
	req.Owner = strings.TrimSpace(model.Owner)
	req.Properties = model.Properties

	remote, err := client.CreateConnection(ctx, req)
	if err != nil {
		return PostgreSQLConnectionModel{}, err
	}

	return mergeConnectionInfo(model, remote), nil
}

func DeletePostgreSQLConnection(ctx context.Context, client dbclient.ConnectionClient, name string) error {
	if client == nil {
		return fmt.Errorf("Databricks connection client is required")
	}
	return client.DeleteConnection(ctx, name)
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
		Owner:                 plan.Owner,
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
	case "comment", "properties", "read_only", "provider_config":
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
		model.ID = remote.Name
		model.Name = remote.Name
	}
	if model.ID == "" {
		model.ID = model.Name
	}
	model.ConnectionID = remote.ID
	model.FullName = remote.FullName
	model.MetastoreID = remote.MetastoreID
	model.CredentialType = remote.CredentialType
	model.URL = remote.URL
	model.CreatedAt = remote.CreatedAt
	model.CreatedBy = remote.CreatedBy
	model.UpdatedAt = remote.UpdatedAt
	model.UpdatedBy = remote.UpdatedBy
	if remote.ProvisioningInfo != nil {
		model.ProvisioningInfo = &ProvisioningInfoModel{State: remote.ProvisioningInfo.State}
	}
	if remote.Comment != "" {
		model.Comment = remote.Comment
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
		model.Owner = remote.Owner
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
