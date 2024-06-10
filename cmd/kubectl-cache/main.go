package main

import (
	"context"
	"log"
	"syscall"

	"github.com/yhlooo/kubectl-cache/pkg/commands"
	"github.com/yhlooo/kubectl-cache/pkg/utils/ctxutil"
)

// Version 版本号
// 构建时注入
var Version = "0.0.0-dev"

func main() {
	// 将信号绑定到上下文
	ctx, cancel := ctxutil.Notify(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	// 创建命令
	cmd := commands.NewRootCommand()
	cmd.Version = Version
	// 执行命令
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
