package grpc_interceptor

import (
	"context"
	"fmt"

	"github.com/goto/compass/core/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserHeaderCtx middleware will propagate a valid user ID as string within request context
// use `user.FromContext` function to get the user ID string
func UserHeaderCtx(identityHeaderKeyEmail string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		userEmail := ""
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", fmt.Errorf("metadata in grpc doesn't exist")
		}

		metadataValues := md.Get(identityHeaderKeyEmail)
		if len(metadataValues) > 0 {
			userEmail = metadataValues[0]
		}

		newCtx := user.NewContext(ctx, user.User{Email: userEmail})
		return handler(newCtx, req)
	}
}
