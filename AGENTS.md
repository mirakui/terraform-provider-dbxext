<!-- SPECKIT START -->
Current feature plan: `specs/001-postgresql-secret-connection/plan.md`

For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan.
<!-- SPECKIT END -->

Follow `.specify/memory/constitution.md` for project governance. Commit
messages, pull request text, review comments, code comments, documentation,
specifications, plans, tasks, and user-facing strings created for this project
MUST be written in English.

Behavior changes MUST be implemented with Red-Green TDD: write or update a
failing test first, run the narrowest practical command to prove RED, implement
the minimal GREEN change, then refactor with relevant tests passing.
