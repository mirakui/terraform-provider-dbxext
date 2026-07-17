package databricks

import (
	"testing"
)

func TestPostgreSQLConnectionMapperBuildsTypedCreateOptions(t *testing.T) {
	t.Parallel()

	req, err := BuildCreateConnectionRequest(PostgreSQLConnectionConfig{
		Name: "psql",
		Host: "postgres.example.com",
		Port: 5432,
		User: "postgres_user",
		PasswordSecret: PasswordSecretReference{
			Scope: "scope",
			Key:   "password",
		},
		PasswordSecretVersion: 1,
	})
	if err != nil {
		t.Fatalf("unexpected mapper error: %v", err)
	}

	if req.ConnectionType != "POSTGRESQL" {
		t.Fatalf("expected POSTGRESQL connection type, got %q", req.ConnectionType)
	}

	expectedOptions := map[string]string{
		"host":     "postgres.example.com",
		"port":     "5432",
		"user":     "postgres_user",
		"password": "secret('scope', 'password')",
	}
	for key, expected := range expectedOptions {
		if req.Options[key] != expected {
			t.Fatalf("expected option %q=%q, got %q", key, expected, req.Options[key])
		}
	}
}

func TestPostgreSQLConnectionUpdateMapperPreservesPasswordSecretReference(t *testing.T) {
	t.Parallel()

	req, err := BuildUpdateConnectionRequest("psql", PostgreSQLConnectionConfig{
		Name: "psql",
		Host: "postgres-new.example.com",
		Port: 5433,
		User: "postgres_user_new",
		PasswordSecret: PasswordSecretReference{
			Scope: "scope",
			Key:   "password",
		},
		PasswordSecretVersion: 2,
		Owner:                 "data-owner@example.com",
		EnvironmentSettings: &EnvironmentSettings{
			EnvironmentVersion: "14.2",
			JavaDependencies:   []string{"org.postgresql:postgresql:42.7.4"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected mapper error: %v", err)
	}

	if req.Name != "psql" {
		t.Fatalf("expected update name psql, got %q", req.Name)
	}
	if req.Owner != "data-owner@example.com" {
		t.Fatalf("expected owner in update request, got %q", req.Owner)
	}
	if req.EnvironmentSettings == nil || req.EnvironmentSettings.EnvironmentVersion != "14.2" {
		t.Fatalf("expected environment settings in update request, got %#v", req.EnvironmentSettings)
	}

	expectedOptions := map[string]string{
		"host":     "postgres-new.example.com",
		"port":     "5433",
		"user":     "postgres_user_new",
		"password": "secret('scope', 'password')",
	}
	for key, expected := range expectedOptions {
		if req.Options[key] != expected {
			t.Fatalf("expected update option %q=%q, got %q", key, expected, req.Options[key])
		}
	}
}

func TestPostgreSQLConnectionRotationMapperReappliesSamePasswordSecretReference(t *testing.T) {
	t.Parallel()

	req, err := BuildPasswordRotationUpdateRequest("psql", PostgreSQLConnectionConfig{
		Name: "psql",
		Host: "postgres.example.com",
		Port: 5432,
		User: "postgres_user",
		PasswordSecret: PasswordSecretReference{
			Scope: "scope",
			Key:   "password",
		},
		PasswordSecretVersion: 2,
	})
	if err != nil {
		t.Fatalf("unexpected mapper error: %v", err)
	}

	if req.Options["password"] != "secret('scope', 'password')" {
		t.Fatalf("expected rotation update to reapply the same password secret reference, got %q", req.Options["password"])
	}
	if req.PasswordSecretVersion != 2 {
		t.Fatalf("expected version marker 2, got %d", req.PasswordSecretVersion)
	}
}

func TestPostgreSQLConnectionMapperPasswordSecretExpressionAcceptsSafeReference(t *testing.T) {
	t.Parallel()

	expr, err := PasswordSecretExpression(PasswordSecretReference{
		Scope: "scope-name",
		Key:   "postgres_password",
	})
	if err != nil {
		t.Fatalf("unexpected expression error: %v", err)
	}

	if expr != "secret('scope-name', 'postgres_password')" {
		t.Fatalf("unexpected secret expression: %q", expr)
	}
}

func TestPostgreSQLConnectionMapperPasswordSecretExpressionRejectsUnsafeReference(t *testing.T) {
	t.Parallel()

	for _, ref := range []PasswordSecretReference{
		{Scope: "", Key: "password"},
		{Scope: "scope", Key: ""},
		{Scope: "bad'scope", Key: "password"},
		{Scope: "scope", Key: "bad'key"},
		{Scope: "bad\\scope", Key: "password"},
		{Scope: "scope", Key: "bad\\key"},
	} {
		t.Run(ref.Scope+"/"+ref.Key, func(t *testing.T) {
			_, err := PasswordSecretExpression(ref)
			if err == nil {
				t.Fatalf("expected secret reference %q/%q to be rejected", ref.Scope, ref.Key)
			}
		})
	}
}
