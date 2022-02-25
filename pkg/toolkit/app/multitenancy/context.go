package multitenancy

import (
	"context"

	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/types/tx"
)

type multitenancyCtxKey string

const (
	UserInfoKey multitenancyCtxKey = "user_info"
)

func WithUserInfo(ctx context.Context, userInfo *UserInfo) context.Context {
	return context.WithValue(ctx, UserInfoKey, userInfo)
}

func UserInfoValue(ctx context.Context) *UserInfo {
	userInfo, ok := ctx.Value(UserInfoKey).(*UserInfo)
	if !ok {
		return nil
	}
	return userInfo
}

func NewContextFromEnvelope(ctx context.Context, envelope *tx.Envelope) context.Context {
	return WithUserInfo(ctx, NewUserInfo(
		envelope.GetHeadersValue(authutils.TenantIDHeader),
		envelope.GetHeadersValue(authutils.UsernameHeader),
	))
}
