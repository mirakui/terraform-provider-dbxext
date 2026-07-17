# Contract: `dbxext_postgresql_connection`

## Resource Summary

`dbxext_postgresql_connection` manages one Databricks Unity Catalog PostgreSQL
external connection with typed Terraform fields and secret-backed password
handling.

The resource fixes the Databricks connection type to `POSTGRESQL`. It does not
expose `connection_type`, `options`, or raw `password` arguments.

## Example

```hcl
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

## Schema Contract

### Required Arguments

| Name | Type | Description |
|------|------|-------------|
| `name` | string | Databricks connection name. |
| `host` | string | PostgreSQL host. |
| `port` | number | PostgreSQL port, 1 through 65535. |
| `user` | string | PostgreSQL username stored as non-secret metadata. |
| `password_secret` | block | Databricks secret reference for the PostgreSQL password. |
| `password_secret_version` | number | Positive integer marker that forces password reference reapplication. |

### `password_secret` Block

| Name | Type | Description |
|------|------|-------------|
| `scope` | string | Databricks secret scope. |
| `key` | string | Databricks secret key. |

### Optional Arguments

| Name | Type | Lifecycle | Description |
|------|------|-----------|-------------|
| `comment` | string | replacement | Free-form Databricks connection comment. |
| `read_only` | bool | replacement | Databricks read-only connection setting. |
| `properties` | map(string) | replacement | Non-secret Databricks connection properties. |
| `owner` | string | in-place update | Databricks connection owner principal. If omitted, the Databricks default owner is read into state. |
| `environment_settings` | block | in-place update | Databricks connection environment settings. |

### `environment_settings` Block

| Name | Type | Description |
|------|------|-------------|
| `environment_version` | string | Optional Databricks connection environment version. |
| `java_dependencies` | list(string) | Optional Java dependency list. |

### Computed Attributes

| Name | Type | Description |
|------|------|-------------|
| `id` | string | Terraform import/state identity. |
| `connection_id` | string | Databricks unique connection identifier. |
| `full_name` | string | Databricks full connection name. |
| `metastore_id` | string | Databricks metastore identifier. |
| `credential_type` | string | Databricks credential type. |
| `url` | string | Databricks-derived remote data source URL. |
| `created_at` | number | Creation timestamp in epoch milliseconds. |
| `created_by` | string | Creator principal. |
| `updated_at` | number | Last update timestamp in epoch milliseconds. |
| `updated_by` | string | Last updater principal. |
| `provisioning_info` | object | Databricks provisioning state details. |

## Validation Contract

- `name`, `host`, `user`, `password_secret.scope`, and `password_secret.key`
  must be non-empty after trimming surrounding whitespace.
- `port` must be an integer from 1 through 65535.
- `password_secret_version` must be a positive integer.
- `password_secret` must appear exactly once.
- Unsupported `connection_type`, `options`, `password`, and `provider_config`
  fields must produce Terraform schema errors.
- Secret scope/key values must be safely representable in a Databricks
  `secret(scope, key)` expression; unsupported values must fail before any
  remote change.
- Resource-level workspace routing metadata is not exposed because the
  Databricks Unity Catalog Connections SDK/API path used by this provider does
  not accept a corresponding field. Users must select the target workspace via
  provider configuration.

## Remote Mapping Contract

Create and update requests send Databricks connection options equivalent to:

| Databricks option | Resource source |
|-------------------|-----------------|
| `host` | `host` |
| `port` | `port` serialized for Databricks |
| `user` | `user` |
| `password` | `secret(password_secret.scope, password_secret.key)` |

The provider must not call any Databricks API that returns the plaintext
password and must not read the secret value.

## Lifecycle Contract

| Change | Expected Terraform action |
|--------|---------------------------|
| `name` | In-place update using Databricks rename support. |
| `host` | In-place update with password secret reference included. |
| `port` | In-place update with password secret reference included. |
| `user` | In-place update with password secret reference included. |
| `password_secret.scope` | In-place update with new password secret reference. |
| `password_secret.key` | In-place update with new password secret reference. |
| `password_secret_version` | In-place update that reapplies the same password secret reference. |
| `owner` | In-place update. |
| `environment_settings` | In-place update. |
| `comment` | Replacement. |
| `properties` | Replacement. |
| `read_only` | Replacement. |

The Databricks create API path used by this provider does not accept `owner`
directly. If `owner` is configured during creation, the provider creates the
connection first and then immediately reconciles the owner with an update that
also sends the configured password secret reference.

## Import Contract

The resource supports importing an existing Databricks connection by the
Databricks connection import identifier used by the upstream provider.

Import must:

- Populate non-secret remote fields that Databricks exposes.
- Never populate a raw password.
- Require user configuration for `user`, `password_secret`, and
  `password_secret_version` before any managed update that needs password
  preservation.

## Diagnostics Contract

Diagnostics must:

- Identify invalid field names and invalid values precisely.
- Identify unusable secret scope/key metadata without revealing secret contents.
- Report Databricks rejection of a secret reference without logging plaintext.
- Avoid printing full remote response bodies if they may contain credential
  material.

## Test Contract

The implementation must include RED/GREEN coverage for:

- Required field validation.
- Port range validation.
- Secret expression generation and escaping/rejection.
- Absence of raw password in state, plan, diagnostics, logs, docs, and examples.
- In-place update for host, port, user, owner, environment settings, and
  password secret metadata.
- Replacement for comment, properties, and read-only.
- Import followed by required user/password metadata configuration.
