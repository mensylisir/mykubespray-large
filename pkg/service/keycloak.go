package service

import (
	"encoding/json"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/utils/keycloak"
	"time"
)

type KeycloakService interface {
	CreateGroup(conf entity.GroupConf) error
	QueryUserByName(conf entity.UserConf) ([]byte, error)
}

type keycloakService struct {
}

func NewKeycloakService() keycloakService {
	return keycloakService{}
}

func (ks keycloakService) CreateGroup(conf entity.GroupConf) error {
	baseConfig := keycloak.BaseConfig{
		BaseUrl:      conf.BaseUrl,
		Reamls:       conf.Reamls,
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
	}
	kconfig := keycloak.KeycloakConfig{
		BaseConfig: baseConfig,
		PasswordConfig: keycloak.PasswordConfig{
			BaseConfig: baseConfig,
			Username:   conf.Username,
			Password:   conf.Password,
		},
		AuthorizationCodeConfig: keycloak.AuthorizationCodeConfig{
			BaseConfig:  baseConfig,
			Code:        conf.Code,
			RedirectURI: conf.RedirectURI,
		},
		ClientCredentialsConfig: keycloak.ClientCredentialsConfig{
			BaseConfig: baseConfig,
		},
		DeviceAuthorizationConfig: keycloak.DeviceAuthorizationConfig{
			BaseConfig: baseConfig,
			DeviceCode: conf.DeviceCode,
		},
	}
	client := keycloak.NewKeycloakClient(conf.GrantType, kconfig, 10*time.Second)
	token, _ := client.GetToken()
	objToken := make(map[string]interface{})
	err := json.Unmarshal([]byte(token), &objToken)
	if err != nil {
		logger.GetLogger().Errorf("Error getting token: %v", err)
		return err
	}
	acessToken := objToken["access_token"].(string)
	group := keycloak.GroupRepresentation{
		Name: conf.Name,
		Path: conf.Path,
	}
	err = client.CreateGroup(acessToken, &group)
	if err != nil {
		logger.GetLogger().Errorf("Creating group failed: %v", err)
		return err
	}
	return nil
}

func (ks keycloakService) QueryUserByName(conf entity.UserConf) ([]byte, error) {
	baseConfig := keycloak.BaseConfig{
		BaseUrl:      conf.BaseUrl,
		Reamls:       conf.Reamls,
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
	}
	kconfig := keycloak.KeycloakConfig{
		BaseConfig: baseConfig,
		ClientCredentialsConfig: keycloak.ClientCredentialsConfig{
			BaseConfig: baseConfig,
		},
	}
	client := keycloak.NewKeycloakClient(conf.GrantType, kconfig, 10*time.Second)
	token, err := client.GetToken()
	if err != nil {
		return nil, err
	}
	objToken := make(map[string]interface{})
	err = json.Unmarshal([]byte(token), &objToken)
	if err != nil {
		logger.GetLogger().Errorf("Error getting token: %v", err)
		return nil, err
	}
	acessToken := objToken["access_token"].(string)
	data, err := client.QueryUserByName(acessToken, conf.Name)
	if err != nil {
		return nil, err
	}
	return data, nil
}
