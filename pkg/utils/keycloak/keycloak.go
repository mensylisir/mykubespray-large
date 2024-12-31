package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/httpx"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	GRANT_TYPE_PASSWORD           = "password"
	GRANT_TYPE_AUTHORIZATION_CODE = "authorization_code"
	GRANT_TYPE_CLIENT_CREDENTIALS = "client_credentials"
	GRANT_TYPE_DEVICE_CODE        = "device_code"
	GRANT_TYPE_OAUTH_DEVICE_CODE  = "urn:ietf:params:oauth:grant-type:device_code"
)

type KeycloakClient interface {
	GetToken() (string, error)
	PostForm(urlStr string, payload url.Values) (string, error)
	CreateUser(token string, user KeycloakUser) error
	DeleteUser(token, userID string) error
	UpdateUser(token, userID string, user KeycloakUser) error
}
type RealmsRepresentation struct {
	Realm   string
	Enabled string
}

type ClientRepresentation struct {
	ClientID     string
	RootURL      string
	RedirectURIs []string
	Protocol     string
	Secret       string
	Enabled      bool
}

type GroupRepresentation struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type UserAttributeRepresentation struct {
	Attributes map[string][]string `json:"attributes"`
}

type ClientMapperRepresentation struct {
	Protocol                  string
	Name                      string
	MapperType                string
	UserAttribute             string
	TokenClaimName            string
	ClaimJsonType             string
	AddToIDToken              bool
	AddToAccessToken          bool
	AddToUserInfo             bool
	Multivalued               bool
	AggregatedAttributeValues bool
}

type BaseConfig struct {
	ClientID     string
	ClientSecret string
	Reamls       string
	BaseUrl      string
	TokenURL     string
	UserURL      string
	ClientURL    string
	GroupURL     string
	AdminURL     string
}

func TokenURL(baseURL, realms string) string {
	return fmt.Sprintf("%s/auth/realms/%s/protocol/openid-connect/token", baseURL, realms)
}

func UserURL(baseURL, realms string) string {
	return fmt.Sprintf("%s/auth/admin/realms/%s/users", baseURL, realms)
}

func ClientURL(baseURL, realms string) string {
	return fmt.Sprintf("%s/auth/admin/realms/%s/clients", baseURL, realms)
}

func GroupURL(baseURL, realms string) string {
	return fmt.Sprintf("%s/auth/admin/realms/%s/groups", baseURL, realms)
}

func AdminURL(baseURL string) string {
	return fmt.Sprintf("%s/auth/admin/realms", baseURL)
}

type PasswordConfig struct {
	BaseConfig BaseConfig
	Username   string
	Password   string
}

type AuthorizationCodeConfig struct {
	BaseConfig  BaseConfig
	Code        string
	RedirectURI string
}

type ClientCredentialsConfig struct {
	BaseConfig BaseConfig
}

type DeviceAuthorizationConfig struct {
	BaseConfig BaseConfig
	DeviceCode string
}

type KeycloakConfig struct {
	BaseConfig                BaseConfig
	PasswordConfig            PasswordConfig
	AuthorizationCodeConfig   AuthorizationCodeConfig
	ClientCredentialsConfig   ClientCredentialsConfig
	DeviceAuthorizationConfig DeviceAuthorizationConfig
}

type keycloakClient struct {
	Config     KeycloakConfig
	GrantType  string
	HTTPClient *http.Client
}

func NewKeycloakClient(grant_type string, config KeycloakConfig, timeout time.Duration) *keycloakClient {
	client := &keycloakClient{
		Config:     *NewKeycloakConfig(grant_type, config),
		GrantType:  grant_type,
		HTTPClient: &http.Client{Timeout: timeout, Transport: &httpx.CustomTransport{}},
	}
	return client
}

func NewKeycloakConfig(grant_type string, config KeycloakConfig) *KeycloakConfig {
	switch grant_type {
	case GRANT_TYPE_PASSWORD:
		kconfig := &KeycloakConfig{
			PasswordConfig: *NewPasswordConfig(config),
		}
		kconfig.BaseConfig = kconfig.PasswordConfig.BaseConfig
		return kconfig
	case GRANT_TYPE_AUTHORIZATION_CODE:
		kconfig := &KeycloakConfig{
			AuthorizationCodeConfig: *NewAuthorizationCodeConfig(config),
		}
		kconfig.BaseConfig = kconfig.AuthorizationCodeConfig.BaseConfig
		return kconfig
	case GRANT_TYPE_CLIENT_CREDENTIALS:
		kconfig := &KeycloakConfig{
			ClientCredentialsConfig: *NewClientCredentialsConfig(config),
		}
		kconfig.BaseConfig = kconfig.ClientCredentialsConfig.BaseConfig
		return kconfig
	case GRANT_TYPE_DEVICE_CODE:
		kconfig := &KeycloakConfig{
			DeviceAuthorizationConfig: *NewDeviceAuthorizationConfig(config),
		}
		kconfig.BaseConfig = kconfig.DeviceAuthorizationConfig.BaseConfig
		return kconfig
	}
	return nil
}

func NewBaseConfig(config KeycloakConfig) *BaseConfig {
	baseClient := BaseConfig{
		ClientID:     config.BaseConfig.ClientID,
		ClientSecret: config.BaseConfig.ClientSecret,
		Reamls:       config.BaseConfig.Reamls,
		BaseUrl:      config.BaseConfig.BaseUrl,
		TokenURL:     TokenURL(config.BaseConfig.BaseUrl, config.BaseConfig.Reamls),
		UserURL:      UserURL(config.BaseConfig.BaseUrl, config.BaseConfig.Reamls),
		ClientURL:    ClientURL(config.BaseConfig.BaseUrl, config.BaseConfig.Reamls),
		GroupURL:     GroupURL(config.BaseConfig.BaseUrl, config.BaseConfig.Reamls),
		AdminURL:     AdminURL(config.BaseConfig.BaseUrl),
	}
	return &baseClient
}

func NewPasswordConfig(config KeycloakConfig) *PasswordConfig {
	config.BaseConfig = config.PasswordConfig.BaseConfig
	passwordConfig := PasswordConfig{
		BaseConfig: *NewBaseConfig(config),
		Username:   config.PasswordConfig.Username,
		Password:   config.PasswordConfig.Password,
	}
	config.PasswordConfig = passwordConfig
	config.BaseConfig = config.PasswordConfig.BaseConfig
	return &passwordConfig
}

func NewAuthorizationCodeConfig(config KeycloakConfig) *AuthorizationCodeConfig {
	config.BaseConfig = config.AuthorizationCodeConfig.BaseConfig
	authorizationCodeConfig := AuthorizationCodeConfig{
		BaseConfig:  *NewBaseConfig(config),
		Code:        config.AuthorizationCodeConfig.Code,
		RedirectURI: config.AuthorizationCodeConfig.RedirectURI,
	}
	config.AuthorizationCodeConfig = authorizationCodeConfig
	config.BaseConfig = config.AuthorizationCodeConfig.BaseConfig
	return &authorizationCodeConfig
}

func NewClientCredentialsConfig(config KeycloakConfig) *ClientCredentialsConfig {
	config.BaseConfig = config.ClientCredentialsConfig.BaseConfig
	clientCredentialsConfig := ClientCredentialsConfig{
		BaseConfig: *NewBaseConfig(config),
	}
	config.ClientCredentialsConfig = clientCredentialsConfig
	config.BaseConfig = config.ClientCredentialsConfig.BaseConfig
	return &clientCredentialsConfig
}

func NewDeviceAuthorizationConfig(config KeycloakConfig) *DeviceAuthorizationConfig {
	config.BaseConfig = config.DeviceAuthorizationConfig.BaseConfig
	deviceAuthorizationConfig := DeviceAuthorizationConfig{
		BaseConfig: *NewBaseConfig(config),
		DeviceCode: config.DeviceAuthorizationConfig.DeviceCode,
	}
	config.DeviceAuthorizationConfig = deviceAuthorizationConfig
	config.BaseConfig = config.DeviceAuthorizationConfig.BaseConfig
	return &deviceAuthorizationConfig
}

func (client *keycloakClient) GetToken() (string, error) {
	payload := url.Values{}
	switch client.GrantType {
	case GRANT_TYPE_PASSWORD:
		payload["grant_type"] = []string{GRANT_TYPE_PASSWORD}
		payload["username"] = []string{client.Config.PasswordConfig.Username}
		payload["password"] = []string{client.Config.PasswordConfig.Password}
		payload["client_id"] = []string{client.Config.BaseConfig.ClientID}
		payload["client_secret"] = []string{client.Config.BaseConfig.ClientSecret}
	case GRANT_TYPE_AUTHORIZATION_CODE:
		payload["grant_type"] = []string{GRANT_TYPE_AUTHORIZATION_CODE}
		payload["code"] = []string{client.Config.AuthorizationCodeConfig.Code}
		payload["redirect_uri"] = []string{client.Config.AuthorizationCodeConfig.RedirectURI}
		payload["client_id"] = []string{client.Config.BaseConfig.ClientID}
		payload["client_secret"] = []string{client.Config.BaseConfig.ClientSecret}
	case GRANT_TYPE_CLIENT_CREDENTIALS:
		payload["grant_type"] = []string{GRANT_TYPE_CLIENT_CREDENTIALS}
		payload["client_id"] = []string{client.Config.BaseConfig.ClientID}
		payload["client_secret"] = []string{client.Config.BaseConfig.ClientSecret}
	case GRANT_TYPE_DEVICE_CODE:
		payload["grant_type"] = []string{GRANT_TYPE_OAUTH_DEVICE_CODE}
		payload["client_id"] = []string{client.Config.BaseConfig.ClientID}
		payload["device_code"] = []string{client.Config.DeviceAuthorizationConfig.DeviceCode}
	}
	return client.PostForm(client.Config.BaseConfig.TokenURL, payload)
}

func (client *keycloakClient) PostForm(urlStr string, payload url.Values) (string, error) {
	resp, err := client.HTTPClient.PostForm(urlStr, payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (client *keycloakClient) CreateUser(token string, user KeycloakUser) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		logger.GetLogger().Errorf("failed to bind user: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", client.Config.BaseConfig.UserURL, bytes.NewBuffer(userJSON))
	if err != nil {
		logger.GetLogger().Errorf("failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		logger.GetLogger().Errorf("failed to create user: %v", resp.StatusCode)
		return fmt.Errorf("failed to create user: %s", resp.Status)
	}

	return nil
}

func (client *keycloakClient) DeleteUser(token, userID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", client.Config.BaseConfig.UserURL, userID), nil)
	if err != nil {
		logger.GetLogger().Errorf("failed to create request: %v", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logger.GetLogger().Errorf("failed to delete user: %v", resp.StatusCode)
		return fmt.Errorf("failed to delete user: %s", resp.Status)
	}

	return nil
}

func (client *keycloakClient) UpdateUser(token, userID string, user KeycloakUser) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", client.Config.BaseConfig.UserURL, userID), bytes.NewBuffer(userJSON))
	if err != nil {
		logger.GetLogger().Errorf("failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logger.GetLogger().Errorf("failed to update user: %v", resp.StatusCode)
		return fmt.Errorf("failed to update user: %s", resp.Status)
	}

	return nil
}

func (client *keycloakClient) CreateRealms(token string, realmsRepresent *RealmsRepresentation) error {
	realmsRepresentation, err := json.Marshal(realmsRepresent)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s", client.Config.BaseConfig.AdminURL), bytes.NewBuffer(realmsRepresentation))
	if err != nil {
		logger.GetLogger().Errorf("failed to create realm: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		logger.GetLogger().Errorf("failed to create realm: %v", resp.StatusCode)
		return fmt.Errorf("failed to  create realm: %s", resp.Status)
	}

	return nil
}

func (client *keycloakClient) CreateClient(token string, clientRepresent *ClientRepresentation) error {
	clientRepresentation, err := json.Marshal(clientRepresent)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s", client.Config.BaseConfig.ClientURL), bytes.NewBuffer(clientRepresentation))
	if err != nil {
		logger.GetLogger().Errorf("failed to create client: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		logger.GetLogger().Errorf("failed to create client: %v", resp.StatusCode)
		return fmt.Errorf("failed to  create client: %s", resp.Status)
	}

	return nil
}

func (client *keycloakClient) CreateGroup(token string, groupRepresent *GroupRepresentation) error {
	groupRepresentation, err := json.Marshal(groupRepresent)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s", client.Config.BaseConfig.GroupURL), bytes.NewBuffer(groupRepresentation))
	if err != nil {
		logger.GetLogger().Errorf("failed to create group: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		logger.GetLogger().Errorf("failed to create group: %v", resp.StatusCode)
		return fmt.Errorf("failed to  create group: %s", resp.Status)
	}

	return nil
}

func (client *keycloakClient) AddUserToGroup(token string, groupID, userID string) error {
	url := fmt.Sprintf("%s/%s/members/%s", client.Config.BaseConfig.GroupURL, groupID, userID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		logger.GetLogger().Errorf("failed to add user to group: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logger.GetLogger().Errorf("failed to add user to group: %v", resp.StatusCode)
		return fmt.Errorf("failed to add user to group: %s", resp.Status)
	}
	return nil
}

func (client *keycloakClient) AddUserAttribute(token, userID string, attributes map[string][]string) error {
	updataData := UserAttributeRepresentation{
		Attributes: attributes,
	}
	data, err := json.Marshal(updataData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", client.Config.BaseConfig.UserURL, userID), bytes.NewBuffer(data))
	if err != nil {
		logger.GetLogger().Errorf("failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logger.GetLogger().Errorf("failed to patch user: %v", resp.StatusCode)
		return fmt.Errorf("failed to patch user: %s", resp.Status)
	}
	return nil
}

func (client *keycloakClient) QueryUserAttribute(token, userID string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", client.Config.BaseConfig.UserURL, userID), nil)
	if err != nil {
		logger.GetLogger().Errorf("failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.GetLogger().Errorf("failed to query user attribute: %v", resp.StatusCode)
		return fmt.Errorf("failed to query user attribute: %s", resp.Status)
	}
	return nil
}

func (client *keycloakClient) QueryUserByName(token, name string) ([]byte, error) {
	url := fmt.Sprintf("%s?username=%s", client.Config.BaseConfig.UserURL, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.GetLogger().Errorf("failed to create request: %v", err)
		return nil, err
	}
	//req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.GetLogger().Errorf("failed to get user: %v", resp.StatusCode)
		return nil, fmt.Errorf("failed to get user: %s", resp.Status)
	}
	responseBody, err := io.ReadAll(resp.Body)
	return responseBody, nil
}

func (client *keycloakClient) AddClientMapper(token, clientID string, representation ClientMapperRepresentation) error {
	data, err := json.Marshal(representation)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/protocol-mappers/models", client.Config.BaseConfig.ClientURL, clientID), bytes.NewBuffer(data))
	if err != nil {
		logger.GetLogger().Errorf("failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		logger.GetLogger().Errorf("failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logger.GetLogger().Errorf("failed to add mapper: %v", resp.StatusCode)
		return fmt.Errorf("failed to add mapper: %s", resp.Status)
	}
	return nil
}
