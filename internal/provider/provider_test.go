package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderSchemaIncludesDatabricksAuthFields(t *testing.T) {
	t.Parallel()

	providerUnderTest := New("test")()
	var resp provider.SchemaResponse

	providerUnderTest.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("provider schema returned diagnostics: %s", resp.Diagnostics.Errors())
	}

	for _, name := range []string{"host", "token"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("expected provider schema attribute %q", name)
		}

		stringAttr, ok := attr.(providerschema.StringAttribute)
		if !ok {
			t.Fatalf("expected %q to be a string attribute, got %T", name, attr)
		}
		if !stringAttr.Optional {
			t.Fatalf("expected %q to be optional so environment variables can be used", name)
		}
	}
}

func TestResolveProviderConfigUsesExplicitValuesBeforeEnv(t *testing.T) {
	t.Setenv("DATABRICKS_HOST", "https://env.example.com")
	t.Setenv("DATABRICKS_TOKEN", "env-token")

	cfg, err := ResolveProviderConfigValues(ProviderConfigInput{
		Host:  "https://explicit.example.com",
		Token: "explicit-token",
	}, os.Getenv)
	if err != nil {
		t.Fatalf("unexpected config resolution error: %v", err)
	}

	if cfg.Host != "https://explicit.example.com" {
		t.Fatalf("expected explicit host, got %q", cfg.Host)
	}
	if cfg.Token != "explicit-token" {
		t.Fatalf("expected explicit token, got %q", cfg.Token)
	}
}

func TestResolveProviderConfigFallsBackToEnv(t *testing.T) {
	t.Setenv("DATABRICKS_HOST", "https://env.example.com")
	t.Setenv("DATABRICKS_TOKEN", "env-token")

	cfg, err := ResolveProviderConfigValues(ProviderConfigInput{}, os.Getenv)
	if err != nil {
		t.Fatalf("unexpected config resolution error: %v", err)
	}

	if cfg.Host != "https://env.example.com" {
		t.Fatalf("expected environment host, got %q", cfg.Host)
	}
	if cfg.Token != "env-token" {
		t.Fatalf("expected environment token, got %q", cfg.Token)
	}
}

func TestResolveProviderConfigRequiresHostAndToken(t *testing.T) {
	t.Parallel()

	_, err := ResolveProviderConfigValues(ProviderConfigInput{}, func(string) string { return "" })
	if err == nil {
		t.Fatal("expected missing host/token to fail")
	}
}

func TestProviderRegistersPostgreSQLConnectionResource(t *testing.T) {
	t.Parallel()

	providerUnderTest := New("test")()
	resources := providerUnderTest.Resources(context.Background())
	if len(resources) != 1 {
		t.Fatalf("expected exactly one resource, got %d", len(resources))
	}

	resourceUnderTest := resources[0]()
	var resp resource.MetadataResponse
	resourceUnderTest.Metadata(context.Background(), resource.MetadataRequest{
		ProviderTypeName: "dbxext",
	}, &resp)

	if resp.TypeName != "dbxext_postgresql_connection" {
		t.Fatalf("expected registered resource type dbxext_postgresql_connection, got %q", resp.TypeName)
	}
}
