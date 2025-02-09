package service

import (
	"context"
	"douyin-mall/internal/payment/model"
	pb "douyin-mall/proto/payment"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PaymentService 支付服务实现
type PaymentService struct {
	pb.UnimplementedPaymentServiceServer
	db *gorm.DB
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{db: db}
}

// Charge 处理支付请求
func (s *PaymentService) Charge(ctx context.Context, req *pb.ChargeReq) (*pb.ChargeResp, error) {
	// 生成唯一交易ID
	transactionID := fmt.Sprintf("TXN-%d", time.Now().UnixNano())

	// 创建支付记录
	payment := model.Payment{
		TransactionID: transactionID,
		UserID:        req.UserId,
		OrderID:       req.OrderId,
		Amount:        float64(req.Amount),
		Status:        "success", // 简化处理，实际应该调用第三方支付
	}

	// 保存支付记录到数据库
	if err := s.db.Create(&payment).Error; err != nil {
		return nil, err
	}

	return &pb.ChargeResp{
		TransactionId: transactionID,
	}, nil
}
