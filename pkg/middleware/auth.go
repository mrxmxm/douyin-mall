package middleware

import (
	"context"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	jwtSecret []byte
}

func NewAuthInterceptor(jwtSecret []byte) *AuthInterceptor {
	return &AuthInterceptor{
		jwtSecret: jwtSecret,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 跳过登录和注册接口的认证
		if info.FullMethod == "/auth.AuthService/Login" || info.FullMethod == "/auth.AuthService/Register" {
			return handler(ctx, req)
		}

		// 从 metadata 中获取 token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		token := md.Get("authorization")
		if len(token) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		// 验证 token
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token[0], claims, func(token *jwt.Token) (interface{}, error) {
			return i.jwtSecret, nil
		})
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// 将用户信息添加到 context
		newCtx := context.WithValue(ctx, "user_id", claims["user_id"])
		newCtx = context.WithValue(newCtx, "email", claims["email"])

		return handler(newCtx, req)
	}
}
