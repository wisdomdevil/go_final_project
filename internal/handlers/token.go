package handlers

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

// token создаёт подписанный токен
func createToken(userPassedPassword string, encKey string) (string, error) {
	// Secret key to sign and verify the token lately
	secret := []byte(encKey)

	signedPassword := HashPassword([]byte(userPassedPassword), secret)

	// создаём payload
	claims := jwt.MapClaims{
		//"password": userPassedPassword, // сюда надо положить захэшированный пароль
		"password": signedPassword, //захэшированный пароль вместе с секретным словом
	}

	// создаём jwt токен и указываем алгоритм хеширования и payload
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// получаем подписанный токен
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		log.Printf("failed to sign jwt: %s\n", err)
		return "", err
	}

	// A [32]byte is not a slice, is a array of 32 bytes,
	// so you need to convert it to a slice using the [:]operator
	// (it means something like "create a slice with all the elements of the array")
	fmt.Println("Result token: " + string(signedToken[:])) // Result token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6WzEyMiw1MywxNzQsMzIsMjAxLDE3NSwyNDIsMjAzLDE0NSwyMzQsMTczLDI0LDExNiwxMTQsNDUsMjYsMjAwLDE5LDcyLDM1LDEzNywxNDMsMTQ1LDE1NCwyNDUsMjQzLDE0MSwxMzYsNTIsOTIsMzEsMTI3XX0.mc1ty-K4zHWjbPJHrpk2nvYMqtocsrUjLGHkpKhfFy4

	return signedToken, nil // string(signedToken[:]), nil
}

// функция для создания подписи
// HashPassword - это hash
func HashPassword(password []byte, secretKey []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(append(password, secretKey...)))
}
