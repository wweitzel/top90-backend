package auth

import "os"

var (
	secretKey = []byte(os.Getenv("TOP90_JWT_SECRET"))
)
