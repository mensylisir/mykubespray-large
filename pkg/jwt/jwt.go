package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"net/http"
	"strings"
)

func VerifyToken(signKey, tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected kwt signing method: %v", token.Header["alg"])
		}
		return []byte(signKey), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractToken(r *http.Request) string {
	tok := r.Header.Get("Authorization")

	if len(tok) > 6 && strings.ToUpper(tok[0:7]) == "BEARER " {
		return tok[7:]
	}

	return ""
}

func ExtractTokenMetadata(r *http.Request) (*entity.KubekeyConf, error) {
	signKey := viper.GetString("jwt.secret")
	token, err := VerifyToken(signKey, ExtractToken(r))
	if err != nil {
		return nil, err
	}

	var user entity.KubekeyConf

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		js, err := json.Marshal(claims)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(js, &user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

	return nil, errors.New("token is invalid")
}
