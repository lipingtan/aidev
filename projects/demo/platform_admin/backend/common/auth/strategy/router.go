package strategy

import "fmt"

// ErrUnsupportedGrantType 不支持的认证类型错误
var ErrUnsupportedGrantType = fmt.Errorf("不支持的认证类型")

// StrategyRouter 策略路由器，根据 grant_type 分发到对应策略
type StrategyRouter struct {
	strategies map[string]AuthenticationStrategy
}

// NewStrategyRouter 创建策略路由器
func NewStrategyRouter() *StrategyRouter {
	return &StrategyRouter{
		strategies: make(map[string]AuthenticationStrategy),
	}
}

// Register 注册认证策略
func (r *StrategyRouter) Register(s AuthenticationStrategy) {
	r.strategies[s.GrantType()] = s
}

// Route 根据 grant_type 查找对应策略
func (r *StrategyRouter) Route(grantType string) (AuthenticationStrategy, error) {
	s, ok := r.strategies[grantType]
	if !ok {
		return nil, ErrUnsupportedGrantType
	}
	return s, nil
}
