package options

import "context"

type globalOptionsContextKey struct{}

// NewContextWithGlobalOptions 将全局选项注入上下文
func NewContextWithGlobalOptions(parent context.Context, opts GlobalOptions) context.Context {
	return context.WithValue(parent, globalOptionsContextKey{}, opts)
}

// GlobalOptionsFromContext 从上下文获取全局选项
func GlobalOptionsFromContext(ctx context.Context) GlobalOptions {
	opts, _ := ctx.Value(globalOptionsContextKey{}).(GlobalOptions)
	return opts
}
