# Research: PostgreSQL Secret-Backed Connection

## Decision: Use Terraform Plugin Framework for the provider scaffold

**Rationale**: The repository does not contain provider source code yet.
HashiCorp documents Terraform Plugin Framework as the recommended SDK for new
providers and current plugin protocol versions. It provides schema attributes,
plan modifiers, import support, diagnostics, and testing patterns needed by the
resource contract.

**Alternatives considered**:
- Terraform Plugin SDKv2: rejected for new provider work because Plugin
  Framework is the recommended path for new providers.
- Wrapping the upstream Databricks provider: rejected because the feature
  changes the public resource schema and password-handling contract.

## Decision: Use Databricks Go SDK catalog connection APIs for CRUD

**Rationale**: The Databricks SDK exposes generated Unity Catalog connection
models with `CreateConnection`, `UpdateConnection`, and `ConnectionInfo` shapes.
Those models include the relevant connection fields: connection type, options,
comment, read-only, owner, properties, environment settings, and URL/readback
metadata. Using the SDK keeps authentication, request construction, and response
typing aligned with Databricks' OpenAPI-generated client.

**Alternatives considered**:
- Handwritten HTTP client: rejected because it duplicates SDK behavior and
  increases compatibility risk.
- SQL statement execution: rejected as the primary path because it would require
  a SQL warehouse dependency that the Unity Catalog Connections API avoids.
  Acceptance tests must still prove that the API accepts Databricks secret
  reference expressions for the password option.

## Decision: Represent password as a Databricks secret expression

**Rationale**: The feature requires the provider to pass a secret reference and
never read the password plaintext. Databricks SQL documentation recommends
credential-bearing connection options use the `secret(scope, key)` function
instead of literal values. The provider will derive the password option value
from `password_secret.scope` and `password_secret.key`, validate the inputs, and
send only the resulting secret reference expression to Databricks.

**Alternatives considered**:
- Read the Databricks secret plaintext and send it transiently: rejected by the
  clarified specification.
- Store password as a Terraform sensitive attribute: rejected because sensitive
  state still persists the raw value.
- Require users to pass a preformatted `secret(...)` string: rejected because it
  reintroduces an untyped option surface and weaker validation.

## Decision: Plain `user`, secret-backed `password`

**Rationale**: The clarified spec requires a plain `user` string. Although
Databricks SQL examples show both `user` and `password` may be credential-bearing
secret expressions, this resource treats username as non-secret metadata and
documents that it can appear in configuration, state, plans, diagnostics, logs,
examples, and generated documentation.

**Alternatives considered**:
- `user_secret` block: rejected by clarification.
- Optional plain or secret username: rejected because the first contract needs a
  single typed path and predictable validation.

## Decision: Lifecycle parity for applicable upstream fields

**Rationale**: Upstream `databricks_connection` documents `comment`,
`properties`, and `read_only` as replacement fields, while the Databricks SDK
update model supports updates for options, owner, new name, and environment
settings. This resource will update typed PostgreSQL options, name, owner,
environment settings, and password secret metadata in place. It will force
replacement for comment, properties, provider configuration, and read-only.

**Alternatives considered**:
- Force replacement for all fields: rejected because it would not address the
  password-preserving update problem.
- In-place update for every field: rejected because upstream lifecycle and SDK
  update support do not cover all fields.

## Decision: Store only secret metadata and version marker

**Rationale**: The resource must track enough information to reapply the remote
password reference when Databricks requires it during updates. Storing
`password_secret.scope`, `password_secret.key`, and `password_secret_version`
provides deterministic planning without storing or reading the plaintext value.

**Alternatives considered**:
- Detect secret value changes automatically: rejected because the provider never
  reads the secret value.
- Store a hash of the secret value: rejected because hashing would require
  reading the plaintext.

## Decision: Import requires follow-up secret metadata before managed update

**Rationale**: Databricks does not expose plaintext credential values on read.
Imported resources can read non-secret connection fields, but users must provide
`user`, `password_secret`, and `password_secret_version` before any update that
requires password preservation.

**Alternatives considered**:
- Import without requiring secret metadata later: rejected because updates could
  erase the password.
- Disallow import entirely: rejected because upstream resources support import
  and operators need adoption workflows.
