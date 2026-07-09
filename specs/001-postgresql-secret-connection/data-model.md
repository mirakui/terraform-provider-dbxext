# Data Model: PostgreSQL Secret-Backed Connection

## Entity: PostgreSQLConnectionResource

Terraform-managed representation of one Databricks Unity Catalog PostgreSQL
external connection.

### Fields

| Field | Type | Required | State | Lifecycle | Validation |
|-------|------|----------|-------|-----------|------------|
| `id` | string | computed | stored | changes after create/import/rename | Databricks connection identifier, derived from remote identity |
| `connection_id` | string | computed | stored | read-only | Non-empty when Databricks returns it |
| `name` | string | yes | stored | in-place rename | Non-empty, unique within metastore |
| `host` | string | yes | stored | in-place update | Non-empty, no surrounding whitespace |
| `port` | number | yes | stored | in-place update | Integer from 1 through 65535 |
| `user` | string | yes | stored | in-place update | Non-empty plain username |
| `password_secret` | object | yes | stored metadata only | in-place update | Exactly one block with `scope` and `key` |
| `password_secret.scope` | string | yes | stored | in-place update | Non-empty, valid for Databricks secret expression generation |
| `password_secret.key` | string | yes | stored | in-place update | Non-empty, valid for Databricks secret expression generation |
| `password_secret_version` | number | yes | stored | in-place update | Positive integer |
| `comment` | string | no | stored | replacement | Optional free-form text |
| `read_only` | bool | no | stored | replacement | Defaults to Databricks/provider default when unset |
| `properties` | map(string) | no | stored | replacement | Keys and values are non-secret strings |
| `owner` | string | no | stored | in-place update | Optional non-empty owner principal |
| `environment_settings` | object | no | stored | in-place update | Mirrors upstream connection environment settings |
| `environment_settings.environment_version` | string | no | stored | in-place update | Optional non-empty string |
| `environment_settings.java_dependencies` | list(string) | no | stored | in-place update | Optional list of non-empty strings |
| `provider_config` | object | no | stored | replacement | Optional workspace routing metadata |
| `provider_config.workspace_id` | number | yes when block set | stored | replacement | Positive workspace identifier |
| `url` | string | computed | stored | read-only | Databricks-derived remote URL |
| `created_at` | number | computed | stored | read-only | Epoch milliseconds |
| `created_by` | string | computed | stored | read-only | Databricks principal |
| `updated_at` | number | computed | stored | read-only | Epoch milliseconds |
| `updated_by` | string | computed | stored | read-only | Databricks principal |
| `metastore_id` | string | computed | stored | read-only | Databricks metastore identifier |
| `full_name` | string | computed | stored | read-only | Databricks full connection name |
| `credential_type` | string | computed | stored | read-only | Databricks credential type |
| `provisioning_info` | object | computed | stored | read-only | Databricks provisioning status |

### Sensitive Data Rules

- The raw PostgreSQL password is not a field in this model.
- `password_secret.scope`, `password_secret.key`, and
  `password_secret_version` are metadata and may be stored.
- The generated Databricks password option value is a secret reference
  expression, not a plaintext password.
- Diagnostics and logs must not include raw password values or remote responses
  that contain credential material.

### Relationships

- `PostgreSQLConnectionResource` owns one `PasswordSecretReference`.
- `PostgreSQLConnectionResource` may include one `EnvironmentSettings` block.
- `PostgreSQLConnectionResource` may include one `ProviderConfig` block.
- `PostgreSQLConnectionResource` maps to one remote Databricks connection with
  fixed type `POSTGRESQL`.

## Entity: PasswordSecretReference

Non-secret metadata identifying a Databricks secret to use for the PostgreSQL
password.

### Fields

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `scope` | string | yes | Non-empty; escaped or rejected if it cannot form a safe Databricks secret expression |
| `key` | string | yes | Non-empty; escaped or rejected if it cannot form a safe Databricks secret expression |

### Derived Value

`password_option = secret('<scope>', '<key>')`

The derivation must escape supported characters safely or reject unsupported
values with a diagnostic before any remote change is attempted.

## Entity: DatabricksConnectionOptions

Internal DTO sent to Databricks for the remote PostgreSQL connection.

### Fields

| Key | Source | Notes |
|-----|--------|-------|
| `host` | `host` | Plain string option |
| `port` | `port` | Serialized as Databricks option value |
| `user` | `user` | Plain string option per clarified contract |
| `password` | `password_secret` | Databricks secret expression only |

## State Transitions

```text
Absent
  -> PlannedCreate
  -> RemoteActive
  -> PlannedInPlaceUpdate
  -> RemoteActive
  -> PlannedReplacement
  -> RemoteActive
  -> PlannedDelete
  -> Absent
```

### Transition Rules

- Create requires name, host, port, user, password secret scope/key, and
  password secret version.
- In-place update is allowed for name, host, port, user, owner, environment
  settings, password secret scope/key, and password secret version.
- Replacement is required for comment, properties, provider configuration, and
  read-only.
- Delete removes the remote connection and clears state.
- Import populates non-secret remote fields and requires configuration of user
  and password secret metadata before managed updates that need password
  preservation.
