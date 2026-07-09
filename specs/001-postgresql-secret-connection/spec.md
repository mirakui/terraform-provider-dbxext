# Feature Specification: PostgreSQL Secret-Backed Connection

**Feature Branch**: `001-postgresql-secret-connection`

**Created**: 2026-07-09

**Status**: Draft

**Input**: User description: "Create a dedicated `dbxext_postgresql_connection`
resource for Databricks external PostgreSQL connections. The resource must avoid
the generic untyped `options` map, enforce required PostgreSQL fields, source the
password from a Databricks secret scope/key, and avoid storing the raw password
in Terraform state while preserving the password during updates."

**Language**: All specification content MUST be written in English.

## Clarifications

### Session 2026-07-09

- Q: How must the provider use the configured Databricks secret when setting the connection password? → A: Provider passes a Databricks secret reference and never reads the secret plaintext.
- Q: Should username be part of the first resource contract? → A: Add a required plain `user` string and support all PostgreSQL-applicable arguments supported by `databricks_connection`.
- Q: Which fields update in place versus force replacement? → A: Typed PostgreSQL options, owner, and environment settings update in place; comment, properties, and read-only force replacement.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create a Typed PostgreSQL Connection (Priority: P1)

A platform engineer defines a PostgreSQL external connection with explicit
fields for the connection name, optional comment, optional read-only mode, host,
port, user, top-level connection metadata, and a password secret reference. They
can create the connection without placing the raw password in configuration or
state.

**Why this priority**: This is the primary value of the feature: a safe,
typed alternative to the generic Databricks connection options map.

**Independent Test**: Apply a valid PostgreSQL connection configuration that
uses a password secret reference, then inspect the managed state and generated
plan output to confirm the raw password is absent while the connection exists.

**Acceptance Scenarios**:

1. **Given** a valid secret scope/key containing the PostgreSQL password,
   **When** the engineer applies a PostgreSQL connection configuration,
   **Then** the connection is created with the configured name, host, port, user,
   comment, read-only setting, and password source without the provider reading
   the secret plaintext.
2. **Given** a connection configuration that omits host, port, user, password
   secret scope, password secret key, or password secret version, **When** the
   engineer validates or plans the configuration, **Then** validation fails
   before any remote connection change is attempted.
3. **Given** a successful connection creation, **When** the engineer inspects
   state, plan output, diagnostics, and examples produced by the provider,
   **Then** the raw password value is not present.

---

### User Story 2 - Update Non-Secret Fields Without Losing Password (Priority: P2)

A platform engineer updates the PostgreSQL connection host, port, user, owner,
or environment settings in place. The provider preserves the configured password
by using the secret reference during the update, preventing credential loss in
Databricks. Changes to comment, properties, or read-only mode are handled as
replacement changes.

**Why this priority**: The feature directly addresses update behavior where
changing a non-secret field can remove the external connection password unless
the password is also supplied.

**Independent Test**: Start from a managed connection with a valid password
secret reference, update only the host, and verify that the connection remains
usable with the same password source after the update.

**Acceptance Scenarios**:

1. **Given** an existing managed PostgreSQL connection, **When** the engineer
   changes only the host and applies the change, **Then** the remote connection
   uses the new host and retains the password from the configured secret.
2. **Given** an existing managed PostgreSQL connection, **When** the engineer
   changes user, owner, or environment settings, **Then** the change is handled
   in place without requiring a raw password value in configuration.
3. **Given** an existing managed PostgreSQL connection, **When** the engineer
   changes comment, properties, or read-only mode, **Then** the plan requires
   replacement and never requires a raw password value in configuration.
4. **Given** the configured password secret reference is rejected by Databricks,
   **When** an update would require password preservation, **Then** the
   operation fails and explains which secret reference could not be used without
   revealing secret contents.

---

### User Story 3 - Rotate Password by Version Marker (Priority: P3)

A platform engineer rotates the value stored in the Databricks secret service
and increments a non-secret version marker in the resource. The provider refreshes
the Databricks connection password from the same scope/key without storing the
password value.

**Why this priority**: Password rotation is required for long-lived
infrastructure, but secret values themselves cannot be safely stored or compared
in Terraform state.

**Independent Test**: Change the secret value outside Terraform, increment the
resource version marker, apply the configuration, and verify that the connection
uses the rotated password while state still contains no raw password.

**Acceptance Scenarios**:

1. **Given** a managed connection and an updated secret value, **When** the
   engineer increments `password_secret_version`, **Then** the provider reapplies
   the password from the configured secret scope/key.
2. **Given** a managed connection and an updated secret value, **When** the
   engineer does not change `password_secret_version`, **Then** no password
   refresh is expected solely from the hidden secret value change.
3. **Given** a changed password secret scope or key, **When** the engineer
   applies the configuration, **Then** the provider uses the new secret reference
   and still avoids exposing the raw password.

---

### Edge Cases

- The configured secret scope does not exist.
- The configured secret key does not exist within an existing scope.
- The active credentials lack permission to reference the configured secret.
- A host or port update is requested while Databricks rejects the password
  secret reference.
- The secret value changes without a corresponding version marker change.
- The password secret scope/key changes at the same time as host or port.
- The user is empty or changed at the same time as host, port, or password
  secret metadata.
- Optional top-level connection metadata differs between configuration and the
  remote connection.
- Comment, properties, or read-only mode changes require replacement.
- The port is empty, non-numeric, lower than 1, or higher than 65535.
- A user attempts to configure a raw password, generic `connection_type`, or
  arbitrary untyped connection option.
- A previously managed connection is read back without exposing credential
  material from Databricks.

## Requirements *(mandatory)*

### Functional Requirements

Functional requirements MUST be testable and MUST describe affected Terraform
provider contract elements when applicable: resources/data sources, schema
fields, validation, lifecycle behavior, diagnostics, import behavior, state
migration impact, and compatibility expectations.

- **FR-001**: The provider MUST expose a dedicated resource named
  `dbxext_postgresql_connection` for Databricks PostgreSQL external
  connections.
- **FR-002**: The resource MUST require a non-empty connection name.
- **FR-003**: The resource MUST support optional comment and read-only settings
  for the connection.
- **FR-004**: The resource MUST support all PostgreSQL-applicable top-level
  arguments supported by `databricks_connection`: comment, read-only setting,
  owner, properties, environment settings, and provider configuration.
- **FR-005**: The resource MUST fix the connection type to PostgreSQL and MUST
  NOT expose a configurable `connection_type` argument.
- **FR-006**: The resource MUST replace the generic `options` map with typed
  PostgreSQL fields for host, port, user, and password secret metadata.
- **FR-007**: The resource MUST require a non-empty host, a valid port in the
  range 1 through 65535, and a non-empty user.
- **FR-008**: The resource MUST require exactly one password secret reference
  containing a non-empty secret scope and non-empty secret key.
- **FR-009**: The resource MUST require a positive integer
  `password_secret_version` marker.
- **FR-010**: Creating the resource MUST create a Databricks PostgreSQL
  connection using the configured name, host, port, user, comment, read-only
  setting, owner, properties, environment settings, provider configuration, and
  password secret reference without reading the secret plaintext.
- **FR-011**: The resource MUST NOT expose a raw password configuration field
  and MUST reject attempts to provide arbitrary untyped connection options.
- **FR-012**: The raw password value MUST NOT be written to Terraform state,
  plan output, diagnostics, logs, generated examples, or documentation.
- **FR-013**: Updating host, port, user, owner, environment settings, password
  secret scope/key, or `password_secret_version` MUST update the existing remote
  connection in place and preserve or refresh the remote connection password
  from the configured secret reference.
- **FR-014**: Changing comment, properties, or read-only mode MUST require
  connection replacement and MUST NOT perform an in-place update.
- **FR-015**: If Databricks rejects the configured password secret reference,
  create or update operations MUST fail and MUST identify the unusable scope/key
  without revealing secret contents.
- **FR-016**: Changing `password_secret_version` MUST cause the provider to
  reapply the password from the configured secret reference even when scope and
  key are unchanged.
- **FR-017**: Reading the resource MUST maintain all non-secret connection
  fields needed for future plans while continuing to hide the raw password.
- **FR-018**: Importing or adopting an existing connection MUST NOT expose the
  raw password and MUST require users to provide user and password secret
  metadata before any managed update that needs password preservation.
- **FR-019**: The initial scope MUST be limited to PostgreSQL external
  connections; other Databricks connection types and arbitrary provider options
  are out of scope.
- **FR-020**: Documentation and examples MUST show typed PostgreSQL fields,
  plain username usage, secret-backed password usage, and PostgreSQL-applicable
  top-level connection metadata, and MUST NOT show literal PostgreSQL passwords.

### Key Entities *(include if feature involves data)*

- **PostgreSQL Connection**: A Databricks external connection identified by name
  and configured with host, port, user, optional comment, optional read-only
  mode, optional owner, optional properties, optional environment settings,
  optional provider configuration, and a password source.
- **Password Secret Reference**: The non-secret metadata that identifies where
  the PostgreSQL password is stored: Databricks secret scope and secret key. The
  provider passes this reference to Databricks and never reads the secret
  plaintext.
- **Password Secret Version Marker**: A user-managed positive integer that
  signals when the hidden password value is to be reapplied from the same secret
  reference.
- **Connection Metadata**: Optional non-secret settings supported by the
  upstream PostgreSQL connection resource, including owner, properties,
  environment settings, and provider configuration.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of configurations missing name, host, port, user, password
  secret scope, password secret key, or password secret version are rejected
  before any remote connection change is attempted.
- **SC-002**: In validation scenarios with a known sample password, state, plan
  output, diagnostics, logs, examples, and generated documentation contain zero
  occurrences of the raw password.
- **SC-003**: In controlled update tests that change only host or port, the
  remote connection retains a usable password in every run.
- **SC-004**: In controlled rotation tests, incrementing the version marker
  reapplies the updated password in every run without requiring a raw password
  in configuration.
- **SC-005**: A platform engineer can define a valid PostgreSQL connection using
  the documented resource fields in under 10 minutes without consulting the
  generic Databricks connection options map.
- **SC-006**: Every PostgreSQL-applicable argument exposed by the upstream
  connection resource is either supported by this resource or documented as not
  applicable because the connection type is fixed and options are typed.
- **SC-007**: 100% of plans that change comment, properties, or read-only mode
  show replacement, while plans that change only typed PostgreSQL options,
  owner, environment settings, or password secret metadata show in-place update.

## Assumptions

- The initial resource contract includes the provided example fields plus a
  required plain `user` field and PostgreSQL-applicable top-level arguments from
  the upstream connection resource.
- The plain `user` value is non-secret metadata and may appear in configuration,
  state, plan output, diagnostics, logs, examples, and documentation.
- Databricks supports using secret references for credential-bearing connection
  options, including PostgreSQL passwords.
- Databricks connection updates require password preservation behavior when
  changing non-secret connection fields, as described in the feature request.
- The provider never reads the raw password value. Password handling is limited
  to passing a Databricks secret reference for Databricks to resolve.
