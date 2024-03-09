package context

import (
	"bao2803/photo_gallery/models"
	"context"
)

type privateKey string

const (
	userKey privateKey = "user"
)

func WithValue(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	if temp := ctx.Value(userKey); temp != nil {
		if user, ok := temp.(*models.User); ok {
			return user
		}
	}
	return nil
}
