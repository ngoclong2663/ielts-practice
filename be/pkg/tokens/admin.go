package tokens

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// AdminJWT issues and verifies admin board tokens. Its own type keeps it apart from the learner
// *JWTManager in dependency injection; its secret is derived from JWT_SECRET, so a learner token
// never verifies as an admin token and vice versa.
type AdminJWT struct {
	*JWTManager
}

func NewAdminJWT(jwtSecret string, ttl time.Duration) *AdminJWT {
	sum := sha256.Sum256([]byte("admin|" + jwtSecret))
	return &AdminJWT{JWTManager: NewJWTManager(hex.EncodeToString(sum[:]), ttl)}
}
