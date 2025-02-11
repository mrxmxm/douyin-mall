package aiclient

import (
	"fmt"
	"strings"
	"time"
)

type Client interface {
	Chat(prompt string) (string, error)
}

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) Chat(prompt string) (string, error) {
	// 模拟处理延迟
	time.Sleep(time.Millisecond * 300)

	// 模拟错误情况
	if strings.Contains(prompt, "error") {
		return "", fmt.Errorf("模拟AI服务错误")
	}

	// 模拟不同场景的回答
	// 商品推荐场景
	if strings.Contains(prompt, "推荐商品") {
		description := strings.TrimPrefix(prompt, "根据描述推荐商品：")
		switch {
		case strings.Contains(description, "手机"):
			return "推荐您购买红米K60，性价比很高，价格实惠，性能强劲。", nil
		case strings.Contains(description, "电脑"):
			return "推荐您购买联想小新Pro16，性能强劲，屏幕出色。", nil
		case strings.Contains(description, "耳机"):
			return "推荐您购买AirPods Pro，降噪效果好，音质出众。", nil
		default:
			return "抱歉，我需要更具体的商品描述才能为您推荐。", nil
		}
	}

	// 订单查询场景
	if strings.Contains(prompt, "订单") {
		switch {
		case strings.Contains(prompt, "状态"):
			return "根据您的订单记录，最近一笔订单正在配送中，预计明天送达。", nil
		case strings.Contains(prompt, "金额"):
			return "您最近的订单金额是999元，包含一部手机和配件。", nil
		case strings.Contains(prompt, "历史"):
			return "您过去30天内共有3笔订单，总金额2999元。", nil
		default:
			return "您有什么具体想了解的订单信息吗？", nil
		}
	}

	// 通用问答场景
	switch {
	case strings.Contains(prompt, "优惠"):
		return "目前平台有多个优惠活动：\n1. 新用户首单立减50元\n2. 满1000减100\n3. 部分商品满2件8折", nil
	case strings.Contains(prompt, "售后"):
		return "商品支持7天无理由退换货，如有质量问题15天内可以申请维修或更换。", nil
	case strings.Contains(prompt, "物流"):
		return "我们的商品一般会在24小时内发货，大部分地区2-3天内送达。", nil
	}

	return "抱歉，我暂时无法理解您的问题。您可以询问订单状态、商品推荐、优惠活动等信息。", nil
}
