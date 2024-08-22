package auth

// LoginService to provide user login with JWT token support
import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	Admin bool
	Exp   time.Time
}

type Claims struct {
	Admin bool `json:"admin"`
	jwt.RegisteredClaims
}

func CreateToken(admin bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"admin": admin,
			"exp":   time.Now().Add(time.Hour * 1).Unix(),
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ReadToken(tokenString string) (Token, error) {
	requestToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return Token{}, err
	}
	if !requestToken.Valid {
		return Token{}, errors.New("invalid token")
	}

	var token Token
	if claims, ok := requestToken.Claims.(*Claims); ok {
		token.Admin = claims.Admin
		token.Exp = claims.RegisteredClaims.ExpiresAt.Time
	}

	return token, nil
}
