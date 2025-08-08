package resolver

import (
	"context"
)

// Context GraphQL resolver上下文
type Context struct {
	context.Context
	// 可以添加用户信息、服务依赖等
}

// NewContext 创建GraphQL resolver上下文
func NewContext(ctx context.Context) *Context {
	return &Context{
		Context: ctx,
	}
}