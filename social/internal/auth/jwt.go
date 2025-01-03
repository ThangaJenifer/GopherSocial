package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret string //Keep this hidden/safe/secure in env file, its a secret
	aud    string
	iss    string
}

func NewJWTAuthenticator(secret, aud, iss string) *JWTAuthenticator {
	return &JWTAuthenticator{secret, iss, aud}
}

// used to generate tokenstring at /authentication/token route in createTokenHandler function handler
// it is present in api/auth.go file
func (a *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	//create a new claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//convert token into string using secret key and need to convert into byte slice
	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}
	return tokenString, err
}

// ex 52 takes in tokenstring and gives jwttoken, used in AuthTokenMiddleware function in api/middleware.go
func (a *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	//using parse method and token function to validate token
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}

		return []byte(a.secret), nil
	},
		//doing some extra checks to make our app more safier
		jwt.WithExpirationRequired(), //first
		jwt.WithAudience(a.aud),      //second
		jwt.WithIssuer(a.aud),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), //highly encouraged and recommended to add
	)

}
