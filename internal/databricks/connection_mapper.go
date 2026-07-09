package databricks

import (
	"fmt"
	"strconv"
	"strings"
)

type PostgreSQLConnectionConfig struct {
	Name                  string
	Host                  string
	Port                  int64
	User                  string
	PasswordSecret        PasswordSecretReference
	PasswordSecretVersion int64
	Owner                 string
	EnvironmentSettings   *EnvironmentSettings
}

type PasswordSecretReference struct {
	Scope string
	Key   string
}

func BuildCreateConnectionRequest(config PostgreSQLConnectionConfig) (ConnectionRequest, error) {
	password, err := PasswordSecretExpression(config.PasswordSecret)
	if err != nil {
		return ConnectionRequest{}, err
	}

	return ConnectionRequest{
		Name:           strings.TrimSpace(config.Name),
		ConnectionType: "POSTGRESQL",
		Options: map[string]string{
			"host":     strings.TrimSpace(config.Host),
			"port":     strconv.FormatInt(config.Port, 10),
			"user":     strings.TrimSpace(config.User),
			"password": password,
		},
		PasswordSecretVersion: config.PasswordSecretVersion,
	}, nil
}

func BuildUpdateConnectionRequest(currentName string, config PostgreSQLConnectionConfig) (ConnectionRequest, error) {
	req, err := buildConnectionRequest(config)
	if err != nil {
		return ConnectionRequest{}, err
	}
	req.Name = strings.TrimSpace(config.Name)
	if req.Name == "" {
		req.Name = strings.TrimSpace(currentName)
	}
	req.Owner = strings.TrimSpace(config.Owner)
	req.EnvironmentSettings = config.EnvironmentSettings
	return req, nil
}

func BuildPasswordRotationUpdateRequest(currentName string, config PostgreSQLConnectionConfig) (ConnectionRequest, error) {
	return BuildUpdateConnectionRequest(currentName, config)
}

func buildConnectionRequest(config PostgreSQLConnectionConfig) (ConnectionRequest, error) {
	password, err := PasswordSecretExpression(config.PasswordSecret)
	if err != nil {
		return ConnectionRequest{}, err
	}

	return ConnectionRequest{
		Name:           strings.TrimSpace(config.Name),
		ConnectionType: "POSTGRESQL",
		Options: map[string]string{
			"host":     strings.TrimSpace(config.Host),
			"port":     strconv.FormatInt(config.Port, 10),
			"user":     strings.TrimSpace(config.User),
			"password": password,
		},
		PasswordSecretVersion: config.PasswordSecretVersion,
	}, nil
}

func PasswordSecretExpression(ref PasswordSecretReference) (string, error) {
	scope := strings.TrimSpace(ref.Scope)
	key := strings.TrimSpace(ref.Key)

	if scope == "" {
		return "", fmt.Errorf("password secret scope is required")
	}
	if key == "" {
		return "", fmt.Errorf("password secret key is required")
	}
	if containsUnsafeSecretExpressionCharacter(scope) {
		return "", fmt.Errorf("password secret scope contains unsupported characters")
	}
	if containsUnsafeSecretExpressionCharacter(key) {
		return "", fmt.Errorf("password secret key contains unsupported characters")
	}

	return fmt.Sprintf("secret('%s', '%s')", scope, key), nil
}

func containsUnsafeSecretExpressionCharacter(value string) bool {
	return strings.ContainsAny(value, "'\\")
}
