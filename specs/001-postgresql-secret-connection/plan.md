# Implementation Plan: PostgreSQL Secret-Backed Connection

**Branch**: `001-postgresql-secret-connection` | **Date**: 2026-07-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-postgresql-secret-connection/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Create the first provider implementation for `terraform-provider-dbxext`: a
typed `dbxext_postgresql_connection` resource that manages Databricks Unity
Catalog PostgreSQL external connections. The resource replaces the upstream
generic `options` map with explicit PostgreSQL fields, passes the password as a
Databricks secret reference, never reads or stores plaintext password material,
and preserves password configuration during in-place updates.

The implementation will scaffold a Go Terraform provider using Terraform Plugin
Framework, use the Databricks Go SDK catalog connection APIs for connection
CRUD, and document/test the public Terraform resource contract before adding
production behavior.

## Technical Context

**Language/Version**: Go, pinned in `go.mod` during implementation with a
version compatible with Terraform Plugin Framework v1.18.x and the selected
Databricks SDK release.

**Primary Dependencies**: Terraform Plugin Framework, Terraform Plugin Testing,
Databricks Go SDK, Terraform Plugin Log, Terraform Plugin Docs.

**Storage**: Terraform state for non-secret metadata only; remote Databricks
Unity Catalog connection stores the actual connection definition. Plaintext
password values are never read by this provider and are never persisted.

**Testing**: `go test ./...`, provider schema/unit tests, plan/state redaction
tests, Terraform Plugin Testing acceptance tests gated behind Databricks
workspace credentials, and documentation generation checks.

**Target Platform**: Terraform CLI managing Databricks workspace-level Unity
Catalog connections.

**Project Type**: Terraform provider plugin.

**Provider Surface**: New resource `dbxext_postgresql_connection`; provider
configuration for Databricks workspace authentication; generated docs and
examples for the resource.

**External APIs**: Databricks Unity Catalog Connections API through the
Databricks Go SDK; Databricks PostgreSQL connection options `host`, `port`,
`user`, and `password`; Databricks `secret(scope, key)` expression for password
credential references.

**Performance Goals**: Planning and validation for a single resource complete
within normal Terraform provider latency; create/update/delete operations
complete within Databricks API latency and surface async provisioning failures.

**Constraints**: Red-Green TDD is mandatory; the provider must never call a
Databricks secret-read API for the password; raw password text must be absent
from state, plans, diagnostics, logs, docs, examples, and tests; comment,
properties, and read-only changes require replacement; typed PostgreSQL options,
owner, environment settings, name, and password secret metadata update in place.
Resource-level workspace routing metadata is not exposed because the Databricks
Unity Catalog Connections SDK/API path used by this provider does not accept a
corresponding field.

**Scale/Scope**: One PostgreSQL connection per resource instance. Initial scope
excludes other connection types and arbitrary untyped options.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Provider contract is defined: `dbxext_postgresql_connection` schema,
      validation, lifecycle, diagnostics, import behavior, state handling, and
      compatibility expectations are explicit in [contracts/dbxext_postgresql_connection.md](./contracts/dbxext_postgresql_connection.md).
- [x] Red-Green TDD is planned: tasks must start with schema/validation and
      redaction tests, then CRUD and lifecycle tests. RED commands are expected
      to use the narrowest `go test` package filters before implementation.
- [x] Lifecycle coverage matches risk: unit tests cover option mapping and
      secret expression escaping; schema tests cover validation/replacement;
      acceptance tests cover Databricks create/update/import/rotation behavior.
- [x] English-language artifacts are planned for specs, plans, tasks,
      documentation, code comments, commit messages, and pull request text.
- [x] Minimal provider design is justified: this repo currently has no provider
      source, so the scaffold is the smallest deployable boundary. The resource
      is limited to PostgreSQL and typed fields.
- [x] Security and state handling are safe: only secret scope/key/version
      metadata is stored; plaintext password values are never retrieved or
      persisted.

Post-design re-check: Passed after `research.md`, `data-model.md`,
`contracts/dbxext_postgresql_connection.md`, and `quickstart.md` were generated.
No constitution gate violations are present.

## Project Structure

### Documentation (this feature)

```text
specs/001-postgresql-secret-connection/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── dbxext_postgresql_connection.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
.
├── main.go
├── go.mod
├── internal/
│   ├── provider/
│   │   ├── provider.go
│   │   ├── provider_test.go
│   │   └── provider_config.go
│   ├── resources/
│   │   ├── postgresql_connection_resource.go
│   │   ├── postgresql_connection_resource_test.go
│   │   └── postgresql_connection_acceptance_test.go
│   └── databricks/
│       ├── client.go
│       ├── connection_mapper.go
│       └── connection_mapper_test.go
├── examples/
│   └── resources/
│       └── dbxext_postgresql_connection/
│           └── resource.tf
├── docs/
│   └── resources/
│       └── postgresql_connection.md
└── tools/
    └── tools.go
```

**Structure Decision**: Use the standard Go Terraform provider layout with
provider wiring in `internal/provider`, resource behavior in
`internal/resources`, Databricks SDK adapters in `internal/databricks`, and
generated documentation/examples under `docs/` and `examples/`. No alternate app
or service layout applies.

## Complexity Tracking

No constitution gate violations require approval. The only notable complexity is
introducing a provider scaffold because this repository currently contains only
Spec Kit artifacts and a feature specification; there is no simpler existing
module boundary to extend.
