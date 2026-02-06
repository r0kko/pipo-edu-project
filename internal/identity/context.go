package identity

import "context"

type ctxKey string

const (
	ctxUserID ctxKey = "user_id"
	ctxRole   ctxKey = "role"
)

func WithUser(ctx context.Context, userID string, role string) context.Context {
	ctx = context.WithValue(ctx, ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxRole, role)
	return ctx
}

func UserIDFrom(ctx context.Context) (string, bool) {
	value := ctx.Value(ctxUserID)
	id, ok := value.(string)
	return id, ok
}

func RoleFrom(ctx context.Context) (string, bool) {
	value := ctx.Value(ctxRole)
	role, ok := value.(string)
	return role, ok
}
