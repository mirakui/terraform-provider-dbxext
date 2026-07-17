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
