package service

import (
	"context"
	"errors"
	"time"

	"douyin-mall/pkg/utils"

	"douyin-mall/proto/auth"
	"douyin-mall/proto/user"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// 验证令牌
func (s *AuthService) VerifyToken(ctx context.Context, req *auth.VerifyTokenRequest) (*auth.VerifyTokenResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return &auth.VerifyTokenResponse{Valid: false}, nil
	}

	claims := token.Claims.(jwt.MapClaims)
	userId := uint32(claims["user_id"].(float64))

	return &auth.VerifyTokenResponse{
		Valid:  true,
		UserId: userId,
	}, nil
}

// 续期令牌
func (s *AuthService) RenewToken(ctx context.Context, req *auth.RenewTokenRequest) (*auth.RenewTokenResponse, error) {
	// 验证旧令牌
	verifyResp, err := s.VerifyToken(ctx, &auth.VerifyTokenRequest{Token: req.OldToken})
	if err != nil || !verifyResp.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// 生成新令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": verifyResp.UserId,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &auth.RenewTokenResponse{
		NewToken: tokenString,
	}, nil
}

// 登出
func (s *AuthService) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	// 可以实现令牌黑名单等逻辑
	return &auth.LogoutResponse{Success: true}, nil
}
