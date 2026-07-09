# Quickstart: PostgreSQL Secret-Backed Connection

This quickstart describes the expected user workflow for the planned
`dbxext_postgresql_connection` resource.

## Prerequisites

- A Databricks workspace with Unity Catalog enabled.
- Credentials that can manage Unity Catalog connections.
- A Databricks secret scope and key containing the PostgreSQL password.
- A PostgreSQL endpoint reachable by Databricks Lakehouse Federation.

## Minimal Configuration

```hcl
terraform {
  required_providers {
    dbxext = {
      source = "local/dbxext"
    }
  }
}

provider "dbxext" {
  host  = var.databricks_host
  token = var.databricks_token
}

resource "dbxext_postgresql_connection" "psql" {
  name      = "psql"
  comment   = "PostgreSQL connection"
  read_only = true

  host = "postgres.example.com"
  port = 5432
  user = "postgres_user"

  password_secret {
    scope = "scope"
    key   = "password"
  }

  password_secret_version = 1
}
```

## Create

1. Store the PostgreSQL password in Databricks Secrets.
2. Configure `password_secret.scope`, `password_secret.key`, and
   `password_secret_version`.
3. Run `terraform plan`.
4. Confirm the plan does not contain the raw password.
5. Run `terraform apply`.
6. Confirm the connection exists in Databricks.

## Update Host or Port

1. Change `host` or `port`.
2. Run `terraform plan`.
3. Confirm Terraform shows an in-place update.
4. Confirm the raw password is absent from the plan.
5. Run `terraform apply`.
6. Confirm the Databricks connection still authenticates with the configured
   password secret reference.

## Rotate Password

1. Replace the secret value stored in Databricks Secrets.
2. Increment `password_secret_version`.
3. Run `terraform plan`.
4. Confirm Terraform shows an in-place update.
5. Confirm the raw password is absent from the plan.
6. Run `terraform apply`.

## Replacement Fields

Changing any of these fields creates a replacement plan:

- `comment`
- `properties`
- `read_only`
- `provider_config`

Replacement must still avoid raw password values in plan output.

## Import

1. Import the existing Databricks connection using the documented import ID.
2. Add `user`, `password_secret`, and `password_secret_version` to the
   configuration before any update that needs password preservation.
3. Run `terraform plan` and confirm no raw password appears.

## Development Validation Commands

```bash
go test ./...
TF_ACC=1 go test ./internal/resources -run TestAccPostgreSQLConnection
go generate ./...
terraform fmt -recursive examples
```

Acceptance tests require Databricks workspace credentials and isolated test
connection names.
