package service

import "context"

type authContextKey struct{}

type AuthContext struct {
	UserID int64
	Email  string
	Role   string
}

func WithAuthContext(ctx context.Context, authCtx AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey{}, authCtx)
}

func AuthFromContext(ctx context.Context) (AuthContext, bool) {
	authCtx, ok := ctx.Value(authContextKey{}).(AuthContext)
	return authCtx, ok
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	authCtx, ok := AuthFromContext(ctx)
	if !ok || authCtx.UserID == 0 {
		return 0, false
	}

	return authCtx.UserID, true
}

func RoleFromContext(ctx context.Context) (string, bool) {
	authCtx, ok := AuthFromContext(ctx)
	if !ok || authCtx.Role == "" {
		return "", false
	}

	return authCtx.Role, true
}
