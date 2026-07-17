package resources_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	dbxextprovider "github.com/mirakui/terraform-provider-dbxext/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccPostgreSQLConnectionResourceName = "dbxext_postgresql_connection.test"

func TestAccPostgreSQLConnectionCreate(t *testing.T) {
	inputs := requirePostgreSQLConnectionAcceptanceInputs(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: postgreSQLConnectionAcceptanceConfig(inputs),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccPostgreSQLConnectionResourceName, "name", inputs.Name),
					resource.TestCheckResourceAttr(testAccPostgreSQLConnectionResourceName, "host", inputs.Host),
					resource.TestCheckResourceAttr(testAccPostgreSQLConnectionResourceName, "port", strconv.FormatInt(inputs.Port, 10)),
					resource.TestCheckResourceAttr(testAccPostgreSQLConnectionResourceName, "user", inputs.User),
					resource.TestCheckResourceAttr(testAccPostgreSQLConnectionResourceName, "password_secret_version", strconv.FormatInt(inputs.PasswordSecretVersion, 10)),
					resource.TestCheckResourceAttrSet(testAccPostgreSQLConnectionResourceName, "connection_id"),
				),
			},
		},
	})
}

func TestAccPostgreSQLConnectionUpdate(t *testing.T) {
	inputs := requirePostgreSQLConnectionAcceptanceInputs(t, "DBXEXT_ACC_POSTGRESQL_UPDATED_HOST")
	updated := inputs
	updated.Host = inputs.UpdatedHost

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: postgreSQLConnectionAcceptanceConfig(inputs),
			},
			{
				Config: postgreSQLConnectionAcceptanceConfig(updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccPostgreSQLConnectionResourceName, "host", updated.Host),
					resource.TestCheckResourceAttr(testAccPostgreSQLConnectionResourceName, "password_secret_version", strconv.FormatInt(updated.PasswordSecretVersion, 10)),
				),
			},
		},
	})
}

func TestAccPostgreSQLConnectionReplacementPlan(t *testing.T) {
	inputs := requirePostgreSQLConnectionAcceptanceInputs(t)
	inputs.Comment = "initial acceptance comment"
	updated := inputs
	updated.Comment = "replacement acceptance comment"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: postgreSQLConnectionAcceptanceConfig(inputs),
			},
			{
				Config:             postgreSQLConnectionAcceptanceConfig(updated),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPostgreSQLConnectionPasswordRotation(t *testing.T) {
	inputs := requirePostgreSQLConnectionAcceptanceInputs(t)
	rotated := inputs
	rotated.PasswordSecretVersion = inputs.PasswordSecretVersion + 1

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: postgreSQLConnectionAcceptanceConfig(inputs),
			},
			{
				Config: postgreSQLConnectionAcceptanceConfig(rotated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccPostgreSQLConnectionResourceName, "password_secret_version", strconv.FormatInt(rotated.PasswordSecretVersion, 10)),
				),
			},
		},
	})
}

func TestPostgreSQLConnectionAcceptanceInputsRequireBaseEnvironment(t *testing.T) {
	t.Parallel()

	_, missing := postgreSQLConnectionAcceptanceInputsFromEnv(func(string) string {
		return ""
	})

	for _, name := range []string{
		"DATABRICKS_HOST",
		"DATABRICKS_TOKEN",
		"DBXEXT_ACC_POSTGRESQL_CONNECTION_NAME",
		"DBXEXT_ACC_POSTGRESQL_HOST",
		"DBXEXT_ACC_POSTGRESQL_PORT",
		"DBXEXT_ACC_POSTGRESQL_USER",
		"DBXEXT_ACC_POSTGRESQL_SECRET_SCOPE",
		"DBXEXT_ACC_POSTGRESQL_SECRET_KEY",
	} {
		if !containsString(missing, name) {
			t.Fatalf("expected missing acceptance input %s, got %#v", name, missing)
		}
	}
}

func TestPostgreSQLConnectionAcceptanceConfigUsesSecretReferenceMetadata(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"DATABRICKS_HOST":                          "https://example.cloud.databricks.com",
		"DATABRICKS_TOKEN":                         "token-placeholder",
		"DBXEXT_ACC_POSTGRESQL_CONNECTION_NAME":    "dbxext_acc_psql",
		"DBXEXT_ACC_POSTGRESQL_HOST":               "postgres.example.com",
		"DBXEXT_ACC_POSTGRESQL_PORT":               "5432",
		"DBXEXT_ACC_POSTGRESQL_USER":               "postgres_user",
		"DBXEXT_ACC_POSTGRESQL_SECRET_SCOPE":       "database",
		"DBXEXT_ACC_POSTGRESQL_SECRET_KEY":         "postgres-password",
		"DBXEXT_ACC_POSTGRESQL_RAW_PASSWORD_DEBUG": "plain-password-sentinel",
	}

	inputs, missing := postgreSQLConnectionAcceptanceInputsFromEnv(func(name string) string {
		return env[name]
	})
	if len(missing) > 0 {
		t.Fatalf("expected complete acceptance inputs, got missing %#v", missing)
	}

	config := postgreSQLConnectionAcceptanceConfig(inputs)

	for _, expected := range []string{
		`resource "dbxext_postgresql_connection" "test"`,
		`name = "dbxext_acc_psql"`,
		`host = "postgres.example.com"`,
		`port = 5432`,
		`user = "postgres_user"`,
		`scope = "database"`,
		`key   = "postgres-password"`,
		`password_secret_version = 1`,
	} {
		if !strings.Contains(config, expected) {
			t.Fatalf("expected generated config to contain %q:\n%s", expected, config)
		}
	}
	if strings.Contains(config, env["DBXEXT_ACC_POSTGRESQL_RAW_PASSWORD_DEBUG"]) {
		t.Fatal("acceptance config must not include raw password material")
	}
	if strings.Contains(config, "provider_config") {
		t.Fatal("acceptance config must not include unsupported provider_config")
	}
}

type postgreSQLConnectionAcceptanceInputs struct {
	Name                  string
	Host                  string
	Port                  int64
	User                  string
	PasswordSecretScope   string
	PasswordSecretKey     string
	PasswordSecretVersion int64
	UpdatedHost           string
	Comment               string
}

func requirePostgreSQLConnectionAcceptanceInputs(t *testing.T, extraRequired ...string) postgreSQLConnectionAcceptanceInputs {
	t.Helper()

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC=1 must be set to run acceptance tests")
	}

	inputs, missing := postgreSQLConnectionAcceptanceInputsFromEnv(os.Getenv)
	for _, name := range extraRequired {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Skipf("acceptance test requires environment variables: %s", strings.Join(missing, ", "))
	}

	return inputs
}

func postgreSQLConnectionAcceptanceInputsFromEnv(getenv func(string) string) (postgreSQLConnectionAcceptanceInputs, []string) {
	required := []string{
		"DATABRICKS_HOST",
		"DATABRICKS_TOKEN",
		"DBXEXT_ACC_POSTGRESQL_CONNECTION_NAME",
		"DBXEXT_ACC_POSTGRESQL_HOST",
		"DBXEXT_ACC_POSTGRESQL_PORT",
		"DBXEXT_ACC_POSTGRESQL_USER",
		"DBXEXT_ACC_POSTGRESQL_SECRET_SCOPE",
		"DBXEXT_ACC_POSTGRESQL_SECRET_KEY",
	}

	values := make(map[string]string, len(required))
	var missing []string
	for _, name := range required {
		values[name] = strings.TrimSpace(getenv(name))
		if values[name] == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return postgreSQLConnectionAcceptanceInputs{}, missing
	}

	port, err := strconv.ParseInt(values["DBXEXT_ACC_POSTGRESQL_PORT"], 10, 64)
	if err != nil || port < 1 || port > 65535 {
		missing = append(missing, "DBXEXT_ACC_POSTGRESQL_PORT")
	}

	passwordSecretVersion := int64(1)
	if raw := strings.TrimSpace(getenv("DBXEXT_ACC_POSTGRESQL_SECRET_VERSION")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 1 {
			missing = append(missing, "DBXEXT_ACC_POSTGRESQL_SECRET_VERSION")
		} else {
			passwordSecretVersion = parsed
		}
	}
	if len(missing) > 0 {
		return postgreSQLConnectionAcceptanceInputs{}, missing
	}

	return postgreSQLConnectionAcceptanceInputs{
		Name:                  values["DBXEXT_ACC_POSTGRESQL_CONNECTION_NAME"],
		Host:                  values["DBXEXT_ACC_POSTGRESQL_HOST"],
		Port:                  port,
		User:                  values["DBXEXT_ACC_POSTGRESQL_USER"],
		PasswordSecretScope:   values["DBXEXT_ACC_POSTGRESQL_SECRET_SCOPE"],
		PasswordSecretKey:     values["DBXEXT_ACC_POSTGRESQL_SECRET_KEY"],
		PasswordSecretVersion: passwordSecretVersion,
		UpdatedHost:           strings.TrimSpace(getenv("DBXEXT_ACC_POSTGRESQL_UPDATED_HOST")),
	}, nil
}

func postgreSQLConnectionAcceptanceConfig(inputs postgreSQLConnectionAcceptanceInputs) string {
	comment := ""
	if strings.TrimSpace(inputs.Comment) != "" {
		comment = fmt.Sprintf("  comment = %s\n", strconv.Quote(strings.TrimSpace(inputs.Comment)))
	}

	return fmt.Sprintf(`
resource "dbxext_postgresql_connection" "test" {
  name = %s
%s
  host = %s
  port = %d
  user = %s

  password_secret {
    scope = %s
    key   = %s
  }

  password_secret_version = %d
}
`,
		strconv.Quote(inputs.Name),
		comment,
		strconv.Quote(inputs.Host),
		inputs.Port,
		strconv.Quote(inputs.User),
		strconv.Quote(inputs.PasswordSecretScope),
		strconv.Quote(inputs.PasswordSecretKey),
		inputs.PasswordSecretVersion,
	)
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"dbxext": providerserver.NewProtocol6WithError(dbxextprovider.New("test")()),
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
