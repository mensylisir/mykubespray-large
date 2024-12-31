package main

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

func main() {
	ParsePrivatekey()
}

func ParsePrivatekey() {
	// 读取私钥文件
	privateKeyFile := "C:\\Users\\Administrator\\workspace\\previous\\work\\work\\workspace\\mykubespray\\pkg\\keycloak\\private_key.pem"
	privateKeyBytes, err := os.ReadFile(privateKeyFile)
	if err != nil {
		fmt.Println("Failed to read private key:", err)
		return
	}

	// 解析私钥
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		fmt.Println("Failed to parse private key:", err)
		return
	}
	expiresIn := time.Now().Add(time.Hour * 24 * 10).Unix()

	claims1 := jwt.MapClaims{
		"exp":           expiresIn,
		"iat":           time.Now().Add(time.Hour * 1).Unix(),
		"auth_time":     0,
		"jti":           "29b10c11-b6a6-446f-9735-45483f253d30",
		"iss":           "https://keycloak.kmpp.io/auth/realms/cars",
		"aud":           "kubernetes",
		"sub":           "4391c288-965f-4d50-adca-499b0ccecb37",
		"typ":           "ID",
		"azp":           "kubernetes",
		"session_state": "9807a7bf-988f-49a6-88b0-1613ebd270bb",
		//"at_hash":            "rF3atus0eQvMvUxD5banYw",
		"acr":                "1",
		"sid":                "9807a7bf-988f-49a6-88b0-1613ebd270aa",
		"email_verified":     false,
		"groups":             []string{"kubernetes-viewer"},
		"name":               "wangzhendong1",
		"preferred_username": "wangzhendong1",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims1)
	token.Header["kid"] = "BMhLqppvHQI_bCAKQOw4JFHcZlwr1mxhJsysaaim4v0"
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		fmt.Println("Failed to sign token:", err)
		return
	}

	fmt.Println("Generated token:", signedToken)
}
