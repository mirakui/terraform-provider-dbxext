# Tasks: PostgreSQL Secret-Backed Connection

**Input**: Design documents from `/specs/001-postgresql-secret-connection/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are mandatory for this behavior change. Use Red-Green TDD:
write or update failing tests first, run the narrowest practical command to
prove RED, implement the minimal GREEN change, then refactor with tests passing.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Provider entrypoint: `main.go`
- Go module and tools: `go.mod`, `go.sum`, `tools/tools.go`
- Provider package: `internal/provider/`
- Databricks adapter package: `internal/databricks/`
- Resource package: `internal/resources/`
- Generated docs and examples: `docs/`, `examples/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the initial Go Terraform provider skeleton and directories.

- [ ] T001 Initialize Go module and dependency pins in go.mod and go.sum
- [ ] T002 Create Terraform provider entrypoint in main.go
- [ ] T003 Create tool dependency pin file in tools/tools.go
- [ ] T004 [P] Create provider package skeleton in internal/provider/provider.go
- [ ] T005 [P] Create provider configuration skeleton in internal/provider/provider_config.go
- [ ] T006 [P] Create Databricks client package skeleton in internal/databricks/client.go
- [ ] T007 [P] Create PostgreSQL connection resource skeleton in internal/resources/postgresql_connection_resource.go
- [ ] T008 [P] Create example directory placeholder in examples/resources/dbxext_postgresql_connection/resource.tf
- [ ] T009 [P] Create generated documentation placeholder in docs/resources/postgresql_connection.md
- [ ] T010 Run `go mod tidy` after setup and update go.mod and go.sum

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish provider wiring, test harnesses, and Databricks client boundaries required by all stories.

**CRITICAL**: No user story work can begin until this phase is complete.

### RED Tests for Foundation

- [ ] T011 [P] Write RED provider configuration tests for host/token/env handling in internal/provider/provider_test.go
- [ ] T012 [P] Write RED Databricks client construction tests in internal/databricks/client_test.go
- [ ] T013 [P] Write RED resource registration test for dbxext_postgresql_connection in internal/provider/provider_test.go
- [ ] T014 Run RED command `go test ./internal/provider ./internal/databricks` and record failure notes in specs/001-postgresql-secret-connection/tasks.md

### GREEN Foundation Implementation

- [ ] T015 Implement provider configuration model and validation in internal/provider/provider_config.go
- [ ] T016 Implement provider factory, metadata, configuration, and resource registration in internal/provider/provider.go
- [ ] T017 Implement Databricks SDK client interface and constructor in internal/databricks/client.go
- [ ] T018 Implement provider entrypoint wiring in main.go
- [ ] T019 Run GREEN command `go test ./internal/provider ./internal/databricks` and record passing notes in specs/001-postgresql-secret-connection/tasks.md

**Checkpoint**: Provider scaffold compiles and can register the resource type.

---

## Phase 3: User Story 1 - Create a Typed PostgreSQL Connection (Priority: P1)

**Goal**: Users can create `dbxext_postgresql_connection` with typed PostgreSQL fields and a password secret reference, without raw password state or logs.

**Independent Test**: Apply or simulate a valid PostgreSQL connection configuration, inspect state/plan/diagnostics/docs/examples, and confirm the raw password is absent while required typed fields are enforced.

### RED Tests for User Story 1

- [ ] T020 [P] [US1] Write RED schema tests for required name, host, port, user, password_secret, and password_secret_version in internal/resources/postgresql_connection_resource_test.go
- [ ] T021 [P] [US1] Write RED schema tests rejecting connection_type, options, and raw password fields in internal/resources/postgresql_connection_resource_test.go
- [ ] T022 [P] [US1] Write RED port and non-empty validation tests in internal/resources/postgresql_connection_resource_test.go
- [ ] T023 [P] [US1] Write RED mapper tests for host/port/user/password option mapping in internal/databricks/connection_mapper_test.go
- [ ] T024 [P] [US1] Write RED secret expression escaping/rejection tests in internal/databricks/connection_mapper_test.go
- [ ] T025 [P] [US1] Write RED create/read/delete resource tests with a mock Databricks client in internal/resources/postgresql_connection_resource_test.go
- [ ] T026 [P] [US1] Write RED acceptance create test skeleton gated by TF_ACC in internal/resources/postgresql_connection_acceptance_test.go
- [ ] T027 [US1] Run RED command `go test ./internal/resources ./internal/databricks -run 'TestPostgreSQLConnection(Create|Schema|Validation|Mapper)'` and record failure notes in specs/001-postgresql-secret-connection/tasks.md

### GREEN Implementation for User Story 1

- [ ] T028 [US1] Implement resource model, schema, validators, and computed fields in internal/resources/postgresql_connection_resource.go
- [ ] T029 [US1] Implement Databricks connection option mapping and secret expression generation in internal/databricks/connection_mapper.go
- [ ] T030 [US1] Implement create/read/delete resource operations using the Databricks client interface in internal/resources/postgresql_connection_resource.go
- [ ] T031 [US1] Implement mock Databricks resource test helpers in internal/resources/postgresql_connection_resource_test.go
- [ ] T032 [US1] Register dbxext_postgresql_connection with the provider in internal/provider/provider.go
- [ ] T033 [US1] Add minimal resource example without raw password in examples/resources/dbxext_postgresql_connection/resource.tf
- [ ] T034 [US1] Add resource documentation for create and validation behavior in docs/resources/postgresql_connection.md
- [ ] T035 [US1] Run GREEN command `go test ./internal/resources ./internal/databricks ./internal/provider -run 'TestPostgreSQLConnection(Create|Schema|Validation|Mapper)'` and record passing notes in specs/001-postgresql-secret-connection/tasks.md

**Checkpoint**: User Story 1 is fully functional and testable independently.

---

## Phase 4: User Story 2 - Update Non-Secret Fields Without Losing Password (Priority: P2)

**Goal**: Users can update in-place fields with the password secret reference preserved, while replacement-only fields plan replacement.

**Independent Test**: Start from a managed connection, update host/port/user/owner/environment settings, and verify the update payload includes the secret reference and contains no plaintext password; verify comment/properties/read_only/provider_config force replacement.

### RED Tests for User Story 2

- [ ] T036 [P] [US2] Write RED update mapper tests for host, port, user, owner, environment_settings, and password secret metadata in internal/databricks/connection_mapper_test.go
- [ ] T037 [P] [US2] Write RED lifecycle plan tests for in-place fields versus replacement fields in internal/resources/postgresql_connection_resource_test.go
- [ ] T038 [P] [US2] Write RED update resource tests with a mock Databricks client in internal/resources/postgresql_connection_resource_test.go
- [ ] T039 [P] [US2] Write RED import/adoption tests requiring user and password secret metadata before managed updates in internal/resources/postgresql_connection_resource_test.go
- [ ] T040 [P] [US2] Write RED acceptance update and replacement-plan skeletons gated by TF_ACC in internal/resources/postgresql_connection_acceptance_test.go
- [ ] T041 [US2] Run RED command `go test ./internal/resources ./internal/databricks -run 'TestPostgreSQLConnection(Update|Lifecycle|Import)'` and record failure notes in specs/001-postgresql-secret-connection/tasks.md

### GREEN Implementation for User Story 2

- [ ] T042 [US2] Implement update request mapping for typed PostgreSQL options, owner, environment_settings, and password secret metadata in internal/databricks/connection_mapper.go
- [ ] T043 [US2] Implement resource Update operation with password secret reference preservation in internal/resources/postgresql_connection_resource.go
- [ ] T044 [US2] Implement plan modifiers requiring replacement for comment, properties, read_only, and provider_config in internal/resources/postgresql_connection_resource.go
- [ ] T045 [US2] Implement owner, properties, environment_settings, and provider_config schema/read mapping in internal/resources/postgresql_connection_resource.go
- [ ] T046 [US2] Implement import state behavior and post-import diagnostics for missing user/password metadata in internal/resources/postgresql_connection_resource.go
- [ ] T047 [US2] Document update, replacement, and import behavior in docs/resources/postgresql_connection.md
- [ ] T048 [US2] Run GREEN command `go test ./internal/resources ./internal/databricks -run 'TestPostgreSQLConnection(Update|Lifecycle|Import)'` and record passing notes in specs/001-postgresql-secret-connection/tasks.md

**Checkpoint**: User Stories 1 and 2 work independently and together.

---

## Phase 5: User Story 3 - Rotate Password by Version Marker (Priority: P3)

**Goal**: Users can rotate a Databricks secret value and increment `password_secret_version` to reapply the secret reference without raw password handling.

**Independent Test**: Change the secret value outside Terraform, increment the version marker, and verify Terraform plans an in-place update that reapplies the secret reference without exposing plaintext.

### RED Tests for User Story 3

- [ ] T049 [P] [US3] Write RED version marker plan tests in internal/resources/postgresql_connection_resource_test.go
- [ ] T050 [P] [US3] Write RED mapper tests proving version-only changes include the same password secret reference in internal/databricks/connection_mapper_test.go
- [ ] T051 [P] [US3] Write RED acceptance rotation skeleton gated by TF_ACC in internal/resources/postgresql_connection_acceptance_test.go
- [ ] T052 [US3] Run RED command `go test ./internal/resources ./internal/databricks -run 'TestPostgreSQLConnection(PasswordSecretVersion|Rotation)'` and record failure notes in specs/001-postgresql-secret-connection/tasks.md

### GREEN Implementation for User Story 3

- [ ] T053 [US3] Implement password_secret_version change handling in internal/resources/postgresql_connection_resource.go
- [ ] T054 [US3] Ensure update mapping reapplies unchanged password secret references on version changes in internal/databricks/connection_mapper.go
- [ ] T055 [US3] Document password rotation workflow in docs/resources/postgresql_connection.md
- [ ] T056 [US3] Run GREEN command `go test ./internal/resources ./internal/databricks -run 'TestPostgreSQLConnection(PasswordSecretVersion|Rotation)'` and record passing notes in specs/001-postgresql-secret-connection/tasks.md

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, generated documentation, examples, and repository hygiene across all stories.

- [ ] T057 [P] Run `go generate ./...` and update docs/resources/postgresql_connection.md
- [ ] T058 [P] Run `terraform fmt -recursive examples` and update examples/resources/dbxext_postgresql_connection/resource.tf
- [ ] T059 [P] Add provider overview and local development commands in README.md
- [ ] T060 [P] Add acceptance-test environment variable documentation in README.md
- [ ] T061 Run full unit test command `go test ./...` and record result in specs/001-postgresql-secret-connection/tasks.md
- [ ] T062 Run `rg -n 'dbxext-raw-password-sentinel' docs/resources/postgresql_connection.md examples/resources/dbxext_postgresql_connection/resource.tf internal/resources/postgresql_connection_resource_test.go internal/databricks/connection_mapper_test.go` and record the expected no-match result in specs/001-postgresql-secret-connection/tasks.md
- [ ] T063 Run acceptance test command `TF_ACC=1 go test ./internal/resources -run TestAccPostgreSQLConnection` when Databricks credentials are available and record result or explicit skip reason in specs/001-postgresql-secret-connection/tasks.md
- [ ] T064 Run `go mod tidy` and verify go.mod and go.sum are stable
- [ ] T065 Review implementation against contracts/dbxext_postgresql_connection.md and update specs/001-postgresql-secret-connection/tasks.md with any resolved deviations

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup completion; blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundation; delivers the MVP create/read/delete contract.
- **User Story 2 (Phase 4)**: Depends on User Story 1 because update/import build on the resource schema and CRUD path.
- **User Story 3 (Phase 5)**: Depends on User Story 2 because password rotation uses update semantics.
- **Polish (Phase 6)**: Depends on all selected user stories.

### User Story Dependencies

- **US1**: Base typed resource creation; required MVP.
- **US2**: Requires US1 schema, mapper, and CRUD behavior.
- **US3**: Requires US2 update behavior.

### Within Each User Story

- RED tests must be written and fail before GREEN implementation begins.
- Mapper tests must precede resource operation implementation.
- Resource schema implementation precedes create/update/import operation implementation.
- Documentation updates follow the implemented behavior they describe.
- Relevant tests must pass before moving to the next story.

---

## Parallel Opportunities

- Setup skeleton tasks T004-T009 can run in parallel after T001-T003 decisions are known.
- Foundation RED tests T011-T013 can run in parallel.
- US1 RED tests T020-T026 can run in parallel because they target separate test concerns.
- US2 RED tests T036-T040 can run in parallel because they target mapper, lifecycle, import, and acceptance concerns separately.
- US3 RED tests T049-T051 can run in parallel.
- Polish documentation, formatting, and README tasks T057-T060 can run in parallel.

## Parallel Example: User Story 1

```bash
# Launch US1 RED tests in parallel-capable work items:
Task: "Write RED schema tests in internal/resources/postgresql_connection_resource_test.go"
Task: "Write RED mapper tests in internal/databricks/connection_mapper_test.go"
Task: "Write RED acceptance create test skeleton in internal/resources/postgresql_connection_acceptance_test.go"
```

## Parallel Example: User Story 2

```bash
# Launch US2 RED tests in parallel-capable work items:
Task: "Write RED update mapper tests in internal/databricks/connection_mapper_test.go"
Task: "Write RED lifecycle plan tests in internal/resources/postgresql_connection_resource_test.go"
Task: "Write RED import tests in internal/resources/postgresql_connection_resource_test.go"
```

## Parallel Example: User Story 3

```bash
# Launch US3 RED tests in parallel-capable work items:
Task: "Write RED version marker tests in internal/resources/postgresql_connection_resource_test.go"
Task: "Write RED mapper tests in internal/databricks/connection_mapper_test.go"
Task: "Write RED acceptance rotation skeleton in internal/resources/postgresql_connection_acceptance_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. Stop and validate create/read/delete behavior, required-field validation, and raw-password absence.

### Incremental Delivery

1. Deliver US1 for typed creation without raw password state.
2. Add US2 for safe in-place updates, replacement lifecycle, and import.
3. Add US3 for password rotation by version marker.
4. Run full polish and acceptance validation.

### TDD Execution Rules

1. For each story, write the listed RED tests first.
2. Run the listed RED command and confirm the failure is caused by missing behavior.
3. Implement the minimum GREEN changes.
4. Run the listed GREEN command.
5. Refactor only with the same tests passing.

## Notes

- [P] tasks use different files or independent test concerns.
- [US1], [US2], and [US3] labels map to the prioritized user stories in spec.md.
- Acceptance tests are gated by `TF_ACC=1` and require isolated Databricks test credentials.
- Never introduce a raw password field or secret-read API call.
- Keep all generated project artifacts in English.
