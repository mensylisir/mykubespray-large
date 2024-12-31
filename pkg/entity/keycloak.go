package entity

import "github.com/whoisfisher/mykubespray/pkg/utils/keycloak"

type BaseKeycloakConf struct {
	BaseUrl      string
	Reamls       string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	Code         string
	RedirectURI  string
	DeviceCode   string
	GrantType    string
}

type GroupConf struct {
	BaseKeycloakConf
	keycloak.GroupRepresentation
}

type UserConf struct {
	BaseKeycloakConf
	Name string
}
