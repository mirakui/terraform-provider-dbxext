package databricks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/catalog"
)

var ErrNotFound = databricks.ErrNotFound

type Config struct {
	Host  string
	Token string
}

type Client interface {
	Connections() ConnectionClient
}

type ConnectionClient interface {
	CreateConnection(ctx context.Context, req ConnectionRequest) (ConnectionInfo, error)
	GetConnection(ctx context.Context, name string) (ConnectionInfo, error)
	UpdateConnection(ctx context.Context, name string, req ConnectionRequest) (ConnectionInfo, error)
	DeleteConnection(ctx context.Context, name string) error
}

type ConnectionRequest struct {
	Name                  string
	ConnectionType        string
	Options               map[string]string
	Comment               *string
	ReadOnly              *bool
	Owner                 string
	Properties            map[string]string
	EnvironmentSettings   *EnvironmentSettings
	PasswordSecretVersion int64
}

type ConnectionInfo struct {
	ID                  string
	Name                string
	FullName            string
	MetastoreID         string
	CredentialType      string
	URL                 string
	CreatedAt           int64
	CreatedBy           string
	UpdatedAt           int64
	UpdatedBy           string
	ConnectionType      string
	Options             map[string]string
	Comment             string
	ReadOnly            *bool
	Owner               string
	Properties          map[string]string
	EnvironmentSettings *EnvironmentSettings
	ProvisioningInfo    *ProvisioningInfo
}

type EnvironmentSettings struct {
	EnvironmentVersion string
	JavaDependencies   []string
}

type ProvisioningInfo struct {
	State string
}

func NewClient(ctx context.Context, cfg Config) (Client, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	workspaceClient, err := databricks.NewWorkspaceClient(&databricks.Config{
		Host:  host,
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("create Databricks workspace client: %w", err)
	}

	return &sdkClient{
		workspace:   workspaceClient,
		connections: &sdkConnectionClient{workspace: workspaceClient},
	}, nil
}

type sdkClient struct {
	workspace   *databricks.WorkspaceClient
	connections *sdkConnectionClient
}

func (c *sdkClient) Connections() ConnectionClient {
	return c.connections
}

type sdkConnectionClient struct {
	workspace *databricks.WorkspaceClient
}

func (c *sdkConnectionClient) CreateConnection(ctx context.Context, req ConnectionRequest) (ConnectionInfo, error) {
	created, err := c.workspace.Connections.Create(ctx, catalog.CreateConnection{
		Name:                req.Name,
		ConnectionType:      catalog.ConnectionTypePostgresql,
		Options:             req.Options,
		Comment:             optionalStringValue(req.Comment),
		ReadOnly:            optionalBoolValue(req.ReadOnly),
		Properties:          req.Properties,
		EnvironmentSettings: toSDKEnvironmentSettings(req.EnvironmentSettings),
		ForceSendFields:     createConnectionForceSendFields(req),
	})
	if err != nil {
		return ConnectionInfo{}, err
	}

	return fromSDKConnectionInfo(created), nil
}

func (c *sdkConnectionClient) GetConnection(ctx context.Context, name string) (ConnectionInfo, error) {
	info, err := c.workspace.Connections.Get(ctx, catalog.GetConnectionRequest{Name: name})
	if err != nil {
		return ConnectionInfo{}, normalizeConnectionError(err)
	}

	return fromSDKConnectionInfo(info), nil
}

func normalizeConnectionError(err error) error {
	if errors.Is(err, databricks.ErrNotFound) || errors.Is(err, databricks.ErrResourceDoesNotExist) {
		return ErrNotFound
	}
	return err
}

func (c *sdkConnectionClient) UpdateConnection(ctx context.Context, name string, req ConnectionRequest) (ConnectionInfo, error) {
	update := catalog.UpdateConnection{
		Name:                name,
		Options:             req.Options,
		Owner:               req.Owner,
		EnvironmentSettings: toSDKEnvironmentSettings(req.EnvironmentSettings),
	}
	if req.Name != "" && req.Name != name {
		update.NewName = req.Name
	}

	updated, err := c.workspace.Connections.Update(ctx, update)
	if err != nil {
		return ConnectionInfo{}, err
	}

	return fromSDKConnectionInfo(updated), nil
}

func (c *sdkConnectionClient) DeleteConnection(ctx context.Context, name string) error {
	return c.workspace.Connections.Delete(ctx, catalog.DeleteConnectionRequest{Name: name})
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func optionalBoolValue(value *bool) bool {
	return value != nil && *value
}

func createConnectionForceSendFields(req ConnectionRequest) []string {
	if req.ReadOnly == nil {
		return nil
	}
	return []string{"ReadOnly"}
}

func toSDKEnvironmentSettings(settings *EnvironmentSettings) *catalog.EnvironmentSettings {
	if settings == nil {
		return nil
	}
	return &catalog.EnvironmentSettings{
		EnvironmentVersion: settings.EnvironmentVersion,
		JavaDependencies:   settings.JavaDependencies,
	}
}

func fromSDKConnectionInfo(info *catalog.ConnectionInfo) ConnectionInfo {
	if info == nil {
		return ConnectionInfo{}
	}

	readOnly := info.ReadOnly
	return ConnectionInfo{
		ID:                  info.ConnectionId,
		Name:                info.Name,
		FullName:            info.FullName,
		MetastoreID:         info.MetastoreId,
		CredentialType:      string(info.CredentialType),
		URL:                 info.Url,
		CreatedAt:           info.CreatedAt,
		CreatedBy:           info.CreatedBy,
		UpdatedAt:           info.UpdatedAt,
		UpdatedBy:           info.UpdatedBy,
		ConnectionType:      string(info.ConnectionType),
		Options:             info.Options,
		Comment:             info.Comment,
		ReadOnly:            &readOnly,
		Owner:               info.Owner,
		Properties:          info.Properties,
		EnvironmentSettings: fromSDKEnvironmentSettings(info.EnvironmentSettings),
		ProvisioningInfo:    fromSDKProvisioningInfo(info.ProvisioningInfo),
	}
}

func fromSDKEnvironmentSettings(settings *catalog.EnvironmentSettings) *EnvironmentSettings {
	if settings == nil {
		return nil
	}
	return &EnvironmentSettings{
		EnvironmentVersion: settings.EnvironmentVersion,
		JavaDependencies:   settings.JavaDependencies,
	}
}

func fromSDKProvisioningInfo(info *catalog.ProvisioningInfo) *ProvisioningInfo {
	if info == nil {
		return nil
	}
	return &ProvisioningInfo{State: string(info.State)}
}
