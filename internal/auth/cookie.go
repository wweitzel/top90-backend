package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
)

func SignCookie(cookie http.Cookie) (http.Cookie, error) {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)
	cookie.Value = string(signature) + cookie.Value
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))
	if len(cookie.String()) > 4096 {
		return http.Cookie{}, errors.New("cookie value too long")
	}
	return cookie, nil
}

func ReadCookie(cookie http.Cookie) (string, error) {
	decodedSignedValue, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", err
	}

	signedValue := string(decodedSignedValue)
	if len(signedValue) < sha256.Size {
		return "", errors.New("invalid cookie length")
	}

	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", errors.New("invalid cookie signature")
	}

	return value, nil
}
