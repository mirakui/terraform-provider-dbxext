<!--
Sync Impact Report
Version change: unversioned template -> 1.0.0
Modified principles:
- Placeholder principles -> I. Terraform Provider Contract First
- Placeholder principles -> II. Red-Green TDD (NON-NEGOTIABLE)
- Placeholder principles -> III. Lifecycle Integration Coverage
- Placeholder principles -> IV. English-Language Project Artifacts
- Placeholder principles -> V. Minimal, Explicit Provider Design
Added sections:
- Additional Constraints
- Development Workflow
Removed sections:
- None
Templates requiring updates:
- ✅ .specify/templates/plan-template.md: updated
- ✅ .specify/templates/spec-template.md: updated
- ✅ .specify/templates/tasks-template.md: updated
- ✅ .specify/templates/checklist-template.md: reviewed, no change required
- ✅ .specify/templates/constitution-template.md: reviewed, no change required
- ✅ .specify/extensions/git/commands/*.md: reviewed, no change required
- ✅ AGENTS.md: updated
Follow-up TODOs:
- None
-->
# Terraform Provider DBXExt Constitution

## Core Principles

### I. Terraform Provider Contract First

All behavior MUST be specified as Terraform provider contract before implementation.
Specs and plans MUST define affected resources/data sources, schema fields,
validation, lifecycle behavior, diagnostics, import behavior, state migration
impact, and compatibility expectations when applicable. Implementation MUST
preserve Terraform plan/apply/read/update/delete semantics and avoid undocumented
side effects.

Rationale: Terraform users depend on predictable plans, stable state, and clear
diagnostics; hidden provider behavior creates drift and unsafe infrastructure
changes.

### II. Red-Green TDD (NON-NEGOTIABLE)

Every behavior change MUST follow Red-Green-Refactor. A failing test MUST be
written or updated first, the RED failure MUST be observed with the narrowest
practical test command, production code MUST then be changed only enough to
reach GREEN, and refactoring MUST keep the relevant tests passing. Exceptions
are allowed only for documentation-only or mechanical metadata changes and MUST
be stated in the plan or pull request.

Rationale: Test-first implementation makes provider semantics explicit before
code changes and prevents accidental acceptance of untested infrastructure
behavior.

### III. Lifecycle Integration Coverage

New or changed provider behavior MUST include coverage at the correct layer:
unit tests for pure logic, contract/schema tests for interfaces, and integration
or acceptance-style tests for Terraform lifecycle behavior, state transitions,
imports, diagnostics, external API edge cases, and regression scenarios. Tests
MUST be isolated, deterministic, and safe to run without leaking credentials or
mutating unrelated infrastructure.

Rationale: Provider defects often appear only across Terraform lifecycle
boundaries; coverage must match the risk surface.

### IV. English-Language Project Artifacts

Commit messages, pull request titles and descriptions, review comments, code
comments, documentation, specifications, plans, tasks, release notes, and
user-facing strings created by this project MUST be written in English.
Identifiers and examples MUST use English unless an external API or protocol
requires another language.

Rationale: A single working language keeps repository history, review context,
generated artifacts, and automation output searchable and maintainable for all
collaborators.

### V. Minimal, Explicit Provider Design

Solutions MUST be scoped to the feature contract and reuse existing project
patterns before adding new abstractions, dependencies, or generated surfaces.
Defaults, retries, pagination, timeouts, state transformations, and breaking
changes MUST be explicit in docs and tests. Any complexity beyond the simplest
viable provider behavior MUST be justified in the plan's Complexity Tracking
section.

Rationale: Terraform providers are long-lived operational tools; small explicit
designs reduce upgrade risk and support burden.

## Additional Constraints

This project is an Apache-2.0 licensed Terraform provider project. Contributions
MUST preserve license notices and MUST not introduce generated or vendored code
without documented source, license, and regeneration steps.

Provider code MUST never log, persist, or expose credentials, tokens, secrets,
or sensitive API responses except through Terraform sensitive attributes designed
for that purpose. Tests and examples MUST use placeholders for secrets and MUST
document any environment variables required for acceptance-style tests.

Feature work MUST keep schema, state, diagnostics, and documentation
synchronized. Any change that affects Terraform configuration, state, import
behavior, or upgrade behavior MUST include migration notes or an explicit
statement that no migration is required.

## Development Workflow

All feature work MUST start from a spec and plan when a Spec Kit feature exists.
The plan MUST complete the Constitution Check before Phase 0 research and repeat
it after Phase 1 design. If no current plan exists, contributors MUST state the
assumption and keep changes narrowly scoped.

Implementation tasks MUST be ordered so RED test tasks precede GREEN
implementation tasks for each user story or behavior slice. Generated `tasks.md`
files MUST include test commands or validation commands that prove RED and GREEN
states. Refactoring, documentation, and review cleanup MUST run relevant tests
again before completion is claimed.

Reviews MUST verify constitution compliance, including English-language
artifacts, Red-Green evidence, test coverage at the correct layer, provider
contract clarity, and documented complexity or migration impact. Pull requests
MUST describe test evidence and any exceptions to this workflow.

## Governance

This constitution supersedes conflicting process guidance in templates, plans,
tasks, local agent instructions, and ad hoc review comments. More specific
technical guidance MAY add constraints, but it MUST not weaken these principles.

Amendments MUST be proposed as documented changes to this file, include a Sync
Impact Report, update dependent templates and runtime guidance in the same
change, and explain the semantic version bump. Ratification requires maintainer
approval through normal repository review.

Versioning follows semantic versioning for governance:
- MAJOR: Removes or redefines a principle, weakens a mandatory rule, or changes
  compliance obligations incompatibly.
- MINOR: Adds a principle or materially expands required workflow, review, or
  testing obligations.
- PATCH: Clarifies wording, fixes typos, or updates non-semantic guidance.

Compliance MUST be checked during planning, task generation, implementation
review, and release preparation. Any approved exception MUST identify the
affected principle, the reason, the risk, and the follow-up action.

**Version**: 1.0.0 | **Ratified**: 2026-07-09 | **Last Amended**: 2026-07-09
