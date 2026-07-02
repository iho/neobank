package auth

import pkgauth "github.com/iho/neobank/pkg/auth"

// JWT wraps shared JWT issuer for the user service.
type JWT = pkgauth.JWT

func NewJWT(secret string) *JWT {
	return pkgauth.NewJWT(secret)
}