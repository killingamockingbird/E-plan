package common

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

func LoadEd25519PrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pemParam, _ := pem.Decode(data)
	if pemParam == nil {
		fmt.Println("fail to decode PEM")
		return nil, err
	}
	key, err := x509.ParsePKCS8PrivateKey(pemParam.Bytes)
	if err != nil {
		return nil, err
	}
	return key.(ed25519.PrivateKey), nil
}

func GetToken(key ed25519.PrivateKey) string {
	claims := jwt.MapClaims{
		"sub": "368BC6DCJ6",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute * 10).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["alg"] = "EdDSA"
	token.Header["kid"] = "K9PRDQY5ET"
	signedToken, err := token.SignedString(key)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	//fmt.Printf("Authorization: %s\n", signedToken)
	return signedToken
}
