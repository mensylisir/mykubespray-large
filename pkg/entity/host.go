package entity

import "golang.org/x/crypto/ssh"

type Host struct {
	Name            string
	Address         string
	InternalAddress string
	User            string
	Password        string
	Port            int32
	Arch            string
	Registry        *Registry
	PrivateKey      string
	AuthMethods     []ssh.AuthMethod
	IsDeleted       bool
}
