---
page_title: "dbxext_postgresql_connection Resource - dbxext"
subcategory: ""
description: |-
  Manages a Databricks PostgreSQL external connection with typed fields and secret-backed password handling.
---

# dbxext_postgresql_connection

`dbxext_postgresql_connection` manages a Databricks Unity Catalog PostgreSQL
external connection. The resource fixes the connection type to PostgreSQL,
uses typed PostgreSQL fields instead of a generic `options` map, and sends the
password as a Databricks secret reference.

The provider stores only `password_secret.scope`, `password_secret.key`, and
`password_secret_version`. It does not read or store the PostgreSQL password
value.

## Example Usage

```terraform
resource "dbxext_postgresql_connection" "psql" {
  name      = "psql"
  comment   = "PostgreSQL connection"
  read_only = true

  host = "postgres.example.com"
  port = 5432
  user = "postgres_user"

  password_secret {
    scope = "database"
    key   = "postgres-password"
  }

  password_secret_version = 1
}
```

## Required

- `name` (String) Databricks connection name.
- `host` (String) PostgreSQL host. Must be non-empty.
- `port` (Number) PostgreSQL port from 1 through 65535.
- `user` (String) PostgreSQL username stored as non-secret metadata.
- `password_secret` (Block) Databricks secret reference for the password.
- `password_secret_version` (Number) Positive marker used to reapply the
  secret reference after secret rotation.

### `password_secret`

- `scope` (String) Databricks secret scope. Must be non-empty and safely
  representable in a Databricks `secret(scope, key)` expression.
- `key` (String) Databricks secret key. Must be non-empty and safely
  representable in a Databricks `secret(scope, key)` expression.

## Optional

- `comment` (String) Databricks connection comment.
- `read_only` (Boolean) Databricks read-only connection setting.
- `owner` (String) Databricks connection owner.
- `properties` (Map of String) Non-secret connection properties.
- `environment_settings` (Block) Databricks connection environment settings.
- `provider_config` (Block) Optional workspace routing metadata.

### `environment_settings`

- `environment_version` (String) Optional Databricks connection environment
  version.
- `java_dependencies` (List of String) Optional Java dependency coordinates.

### `provider_config`

- `workspace_id` (Number) Workspace ID for account-provider-managed resources.
  Changing this block requires replacement.

## Computed

- `id` (String) Terraform state identity.
- `connection_id` (String) Databricks connection identifier.
- `full_name` (String) Databricks full connection name.
- `metastore_id` (String) Databricks metastore identifier.
- `credential_type` (String) Databricks credential type.
- `url` (String) Databricks-derived remote data source URL.
- `created_at` (Number) Creation timestamp in epoch milliseconds.
- `created_by` (String) Creator principal.
- `updated_at` (Number) Last update timestamp in epoch milliseconds.
- `updated_by` (String) Last updater principal.
- `provisioning_info` (Object) Databricks provisioning status.
  - `state` (String) Databricks provisioning state.

## Validation

The provider rejects empty `name`, `host`, `user`, `password_secret.scope`,
and `password_secret.key` values before calling Databricks. It also rejects
ports outside `1` through `65535`, non-positive `password_secret_version`
values, and secret scope/key values that cannot be safely represented in the
Databricks secret expression.

The resource does not expose `connection_type`, `options`, or raw `password`
arguments.

## Update Behavior

Changes to these fields update the existing Databricks connection in place:

- `name`
- `host`
- `port`
- `user`
- `owner`
- `environment_settings`
- `password_secret`
- `password_secret_version`

Every in-place update sends the configured password secret reference again so
Databricks does not drop the existing external connection password while
updating non-secret fields.

Changes to these fields require replacement:

- `comment`
- `properties`
- `read_only`
- `provider_config`

## Import

Import uses the Databricks connection name. Imported resources do not contain a
raw password. Before applying any managed update after import, configure
`user`, `password_secret`, and `password_secret_version` so the provider can
preserve or refresh the password with the Databricks secret reference.

## Password Rotation

To rotate the PostgreSQL password, update the Databricks secret value outside
Terraform and increment `password_secret_version`. The next apply performs an
in-place update and sends the same `password_secret.scope` and
`password_secret.key` reference again. If the hidden secret value changes but
`password_secret_version` does not change, Terraform has no password value to
compare and does not refresh the connection solely from that hidden change.
