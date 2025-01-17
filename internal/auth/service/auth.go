package service

import (
	"context"
	"errors"
	"time"

	"douyin-mall/pkg/utils"

	"douyin-mall/proto/auth"
	"douyin-mall/proto/user"

	"github.com/golang-jwt/jwt"
)

type AuthService struct {
	userClient user.UserServiceClient
	jwtSecret  []byte
	auth.UnimplementedAuthServiceServer
}

func NewAuthService(userClient user.UserServiceClient) *AuthService {
	return &AuthService{
		userClient: userClient,
		jwtSecret:  []byte("your-secret-key"), // 建议从配置文件读取
	}
}

func (s *AuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// 通过 RPC 调用用户服务获取用户信息
	resp, err := s.userClient.GetUserByEmail(ctx, &user.GetUserByEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		return nil, err
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, resp.Password) {
		return nil, errors.New("invalid password")
	}

	// 生成 JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = resp.Id
	claims["email"] = resp.Email
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponse{
		Token: tokenString,
	}, nil
}
