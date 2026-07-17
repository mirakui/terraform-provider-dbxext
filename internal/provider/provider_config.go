package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProviderConfig struct {
	Host  string
	Token string
}

type ProviderConfigInput struct {
	Host  string
	Token string
}

func ResolveProviderConfigValues(input ProviderConfigInput, getenv func(string) string) (ProviderConfig, error) {
	host := strings.TrimSpace(input.Host)
	if host == "" {
		host = strings.TrimSpace(getenv("DATABRICKS_HOST"))
	}

	token := strings.TrimSpace(input.Token)
	if token == "" {
		token = strings.TrimSpace(getenv("DATABRICKS_TOKEN"))
	}

	var missing []string
	if host == "" {
		missing = append(missing, "host")
	}
	if token == "" {
		missing = append(missing, "token")
	}
	if len(missing) > 0 {
		return ProviderConfig{}, fmt.Errorf("missing Databricks provider configuration: %s", strings.Join(missing, ", "))
	}

	return ProviderConfig{
		Host:  host,
		Token: token,
	}, nil
}

func ResolveProviderConfig(ctx context.Context, req provider.ConfigureRequest) (ProviderConfig, error) {
	var config providerConfigModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		return ProviderConfig{}, fmt.Errorf("invalid provider configuration")
	}

	var input ProviderConfigInput
	if !config.Host.IsNull() && !config.Host.IsUnknown() {
		input.Host = config.Host.ValueString()
	}
	if !config.Token.IsNull() && !config.Token.IsUnknown() {
		input.Token = config.Token.ValueString()
	}

	return ResolveProviderConfigValues(input, getenv)
}

type providerConfigModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

var getenv = func(name string) string {
	return os.Getenv(name)
}
