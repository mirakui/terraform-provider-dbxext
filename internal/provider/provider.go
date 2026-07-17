package provider

import (
	"context"

	dbclient "github.com/mirakui/terraform-provider-dbxext/internal/databricks"
	"github.com/mirakui/terraform-provider-dbxext/internal/resources"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = (*DBXExtProvider)(nil)

type DBXExtProvider struct {
	version   string
	newClient func(context.Context, dbclient.Config) (dbclient.Client, error)
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DBXExtProvider{
			version:   version,
			newClient: dbclient.NewClient,
		}
	}
}

func (p *DBXExtProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dbxext"
	resp.Version = p.version
}

func (p *DBXExtProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Configure the DBXExt provider with Databricks workspace authentication.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Databricks workspace URL. May also be set with DATABRICKS_HOST.",
			},
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Databricks personal access token. May also be set with DATABRICKS_TOKEN.",
			},
		},
	}
}

func (p *DBXExtProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	cfg, err := ResolveProviderConfig(ctx, req)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Databricks provider configuration", err.Error())
		return
	}

	client, err := p.newClient(ctx, dbclient.Config{
		Host:  cfg.Host,
		Token: cfg.Token,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to configure Databricks client", err.Error())
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *DBXExtProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewPostgreSQLConnectionResource,
	}
}

func (p *DBXExtProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}
