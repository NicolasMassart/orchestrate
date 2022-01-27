package jwt

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=validator.go -destination=mock/validator.go -package=mock

type Validator interface {
	ValidateToken(ctx context.Context, token string) (*entities.UserClaims, error)
}
