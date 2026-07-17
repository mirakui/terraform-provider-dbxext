# terraform-provider-dbxext

`terraform-provider-dbxext` extends Databricks Terraform workflows with typed
resources for operational cases that are hard to model safely with generic
provider arguments.

## Resource: `dbxext_postgresql_connection`

`dbxext_postgresql_connection` manages a Databricks Unity Catalog PostgreSQL
external connection. It exposes typed PostgreSQL fields for `host`, `port`, and
`user`, fixes the Databricks connection type to PostgreSQL, and sets the remote
password from a Databricks secret reference:

```hcl
password_secret {
  scope = "database"
  key   = "postgres-password"
}
```

The provider stores only the secret scope, key, and version marker. It does not
read or store the raw PostgreSQL password.

## Local Development

```bash
go test ./...
go generate ./...
terraform fmt -recursive examples
```

The implementation uses Terraform Plugin Framework and the Databricks Go SDK.
When running in a sandbox, set writable Go caches if the default Go cache is not
available:

```bash
GOCACHE=/private/tmp/dbxext-gocache GOMODCACHE=/private/tmp/dbxext-gomodcache go test ./...
```

## Acceptance Tests

Acceptance tests are gated by `TF_ACC=1` and require Databricks workspace
credentials with permission to manage Unity Catalog connections.

Expected environment variables:

```bash
export DATABRICKS_HOST="https://example.cloud.databricks.com"
export DATABRICKS_TOKEN="..."
export DBXEXT_ACC_POSTGRESQL_CONNECTION_NAME="dbxext_acc_psql"
export DBXEXT_ACC_POSTGRESQL_HOST="postgres.example.com"
export DBXEXT_ACC_POSTGRESQL_PORT="5432"
export DBXEXT_ACC_POSTGRESQL_USER="postgres_user"
export DBXEXT_ACC_POSTGRESQL_SECRET_SCOPE="database"
export DBXEXT_ACC_POSTGRESQL_SECRET_KEY="postgres-password"
export TF_ACC=1
```

Run the PostgreSQL connection acceptance tests with:

```bash
TF_ACC=1 go test ./internal/resources -run TestAccPostgreSQLConnection
```

The update acceptance test additionally requires
`DBXEXT_ACC_POSTGRESQL_UPDATED_HOST` so the test can verify a host change while
reapplying the same password secret reference. Use isolated connection names
and a Databricks secret scope/key containing the PostgreSQL password when
enabling full acceptance scenarios.
