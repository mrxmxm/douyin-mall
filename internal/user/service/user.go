package service

import (
	"context"
	"errors"

	"douyin-mall/internal/user/model"
	"douyin-mall/pkg/utils"

	pb "douyin-mall/proto/user"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
	pb.UnimplementedUserServiceServer
}

type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// Login 用户登录
func (s *UserService) LoginHTTP(ctx context.Context, c *app.RequestContext) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{
			"code": 400,
			"msg":  "invalid request",
		})
		return
	}

	// 查找用户
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{
			"code": 401,
			"msg":  "invalid email or password",
		})
		return
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{
			"code": 401,
			"msg":  "invalid email or password",
		})
		return
	}

	// TODO: 调用认证服务生成 token
	// 这里需要实现 RPC 调用认证服务

	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": 200,
		"msg":  "login success",
		"data": map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	})
}

func (s *UserService) GetUserByID(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	var u model.User
	if err := s.db.First(&u, req.Id).Error; err != nil {
		return nil, err
	}
	return &pb.UserResponse{
		Id:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.UserResponse, error) {
	var u model.User
	if err := s.db.Where("email = ?", req.Email).First(&u).Error; err != nil {
		return nil, err
	}
	return &pb.UserResponse{
		Id:        u.ID,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}, nil
}

func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// 验证密码
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("passwords do not match")
	}

	// 检查邮箱是否已存在
	var existUser model.User
	if err := s.db.Where("email = ?", req.Email).First(&existUser).Error; err == nil {
		return nil, errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("database error")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("password encryption failed")
	}

	// 创建用户
	user := model.User{
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("failed to create user")
	}

	return &pb.RegisterResponse{
		UserId: user.ID,
	}, nil
}

func (s *UserService) RegisterHTTP(ctx context.Context, c *app.RequestContext) {
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{
			"code": 400,
			"msg":  "invalid request",
		})
		return
	}

	// 调用 gRPC 方法
	resp, err := s.Register(ctx, &pb.RegisterRequest{
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	})

	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": 200,
		"msg":  "register success",
		"data": map[string]interface{}{
			"user_id": resp.UserId,
		},
	})
}
