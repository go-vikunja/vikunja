package websocket

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/golang-jwt/jwt/v5"
)

const authTypeLinkShare = 1

// ValidateToken validates a JWT token and returns the user ID if valid.
// Returns 0 and an error if the token is invalid.
func ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(config.ServiceJWTSecret.GetString()), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, jwt.ErrTokenInvalidClaims
	}

	// Check token type - only allow regular user tokens, not link shares
	typ, ok := claims["type"].(float64)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}
	if int(typ) == authTypeLinkShare {
		return 0, jwt.ErrTokenInvalidClaims
	}

	// Check for API token
	if tokenID, ok := claims["token_id"].(float64); ok && tokenID > 0 {
		return validateAPIToken(tokenString, claims)
	}

	// Get user ID from claims
	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	return int64(userIDFloat), nil
}

func validateAPIToken(tokenString string, _ jwt.MapClaims) (int64, error) {
	s := db.NewSession()
	defer s.Close()

	token, err := models.GetTokenFromTokenString(s, tokenString)
	if err != nil {
		return 0, err
	}

	u, err := user.GetUserByID(s, token.OwnerID)
	if err != nil {
		return 0, err
	}

	return u.ID, nil
}
