package middleware

import (
	"context"
	"strings"

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
		// 白名单路径直接放行
		if isWhitelistPath(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		token := md.Get("authorization")
		if len(token) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		// 验证用户权限
		claims, err := i.validateToken(token[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// 检查用户是否有权限访问该接口
		if !hasPermission(claims["role"].(string), info.FullMethod) {
			return nil, status.Error(codes.PermissionDenied, "no permission")
		}

		return handler(ctx, req)
	}
}

// 白名单路径
var whitelistPaths = []string{
	"/auth.AuthService/Login",
	"/auth.AuthService/Register",
}

func isWhitelistPath(path string) bool {
	for _, p := range whitelistPaths {
		if p == path {
			return true
		}
	}
	return false
}

func (i *AuthInterceptor) validateToken(tokenStr string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return i.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}

func hasPermission(role string, method string) bool {
	// 管理员可以访问所有接口
	if role == "admin" {
		return true
	}

	// 普通用户只能访问基础接口
	if role == "user" {
		return !strings.HasPrefix(method, "/admin.")
	}

	return false
}
