package service

import (
	"context"
	"douyin-mall/internal/product/model"
	"douyin-mall/proto/product"
	"strings"

	"gorm.io/gorm"
)

type ProductService struct {
	product.UnimplementedProductCatalogServiceServer
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

func (s *ProductService) ListProducts(ctx context.Context, req *product.ListProductsReq) (*product.ListProductsResp, error) {
	var products []model.Product
	query := s.db

	if req.CategoryName != "" {
		query = query.Where("categories LIKE ?", "%"+req.CategoryName+"%")
	}

	offset := (req.Page - 1) * int32(req.PageSize)
	if err := query.Offset(int(offset)).Limit(int(req.PageSize)).Find(&products).Error; err != nil {
		return nil, err
	}

	resp := &product.ListProductsResp{
		Products: make([]*product.Product, 0, len(products)),
	}

	for _, p := range products {
		resp.Products = append(resp.Products, &product.Product{
			Id:          uint32(p.ID),
			Name:        p.Name,
			Description: p.Description,
			Picture:     p.Picture,
			Price:       float32(p.Price),
			Categories:  strings.Split(p.Categories, ","),
		})
	}

	return resp, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *product.GetProductReq) (*product.GetProductResp, error) {
	var p model.Product
	if err := s.db.First(&p, req.Id).Error; err != nil {
		return nil, err
	}

	return &product.GetProductResp{
		Product: &product.Product{
			Id:          uint32(p.ID),
			Name:        p.Name,
			Description: p.Description,
			Picture:     p.Picture,
			Price:       float32(p.Price),
			Categories:  strings.Split(p.Categories, ","),
		},
	}, nil
}

func (s *ProductService) SearchProducts(ctx context.Context, req *product.SearchProductsReq) (*product.SearchProductsResp, error) {
	var products []model.Product

	// 简单的模糊查询实现
	if err := s.db.Where(
		"name LIKE ? OR description LIKE ?",
		"%"+req.Query+"%",
		"%"+req.Query+"%",
	).Find(&products).Error; err != nil {
		return nil, err
	}

	resp := &product.SearchProductsResp{
		Results: make([]*product.Product, 0, len(products)),
	}

	for _, p := range products {
		resp.Results = append(resp.Results, &product.Product{
			Id:          uint32(p.ID),
			Name:        p.Name,
			Description: p.Description,
			Picture:     p.Picture,
			Price:       float32(p.Price),
			Categories:  strings.Split(p.Categories, ","),
		})
	}

	return resp, nil
}

// 实现其他接口...
