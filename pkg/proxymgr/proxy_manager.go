package proxymgr

import (
	"context"

	"k8s.io/client-go/rest"
)

// ProxyManager 代理服务管理器
type ProxyManager interface {
	// List 列出所有正在运行的代理
	List(ctx context.Context) ([]ProxyInfo, error)
	// GetForConfig 获取使用指定客户端配置的代理
	GetForConfig(ctx context.Context, config *rest.Config) (*ProxyInfo, error)
	// NewForConfig 使用指定客户端配置创建一个代理
	NewForConfig(ctx context.Context, config *rest.Config) (*ProxyInfo, error)

	// LockConfig 当前进程认领并锁定客户端配置对应的代理（避免其它进程基于此客户端配置启动代理）
	LockConfig(ctx context.Context, config *rest.Config) (*ProxyInfo, error)
	// UnlockConfig 解锁当前进程认领的客户端配置对应的代理
	UnlockConfig(ctx context.Context, config *rest.Config) error
	// SetProxyInfoForConfig 设置客户端配置对应的代理信息
	// NOTE: 仅能设置当前进程提供的代理服务信息，需要先 LockConfig
	SetProxyInfoForConfig(ctx context.Context, config *rest.Config, info *ProxyInfo) error
}

// ProxyInfo 代理信息
type ProxyInfo struct {
	// 代理服务进程 ID
	PID int

	// 代理服务端口
	Port int
	// 代理服务 UNIX Socket 路径
	UNIXSocketPath string

	// 代理服务的客户端配置签名
	ClientConfigSignature string
}

// NewProxyManager 创建一个代理服务管理器
func NewProxyManager(dataRoot string) ProxyManager {
	return &defaultProxyManager{dataRoot: dataRoot}
}

// defaultProxyManager 是 ProxyManager 的一个默认实现
type defaultProxyManager struct {
	dataRoot string
}

var _ ProxyManager = &defaultProxyManager{}

// List 列出所有正在运行的代理
func (mgr *defaultProxyManager) List(ctx context.Context) ([]ProxyInfo, error) {
	//TODO implement me
	panic("implement me")
}

// GetForConfig 获取使用指定客户端配置的代理
func (mgr *defaultProxyManager) GetForConfig(ctx context.Context, config *rest.Config) (*ProxyInfo, error) {
	//TODO implement me
	panic("implement me")
}

// NewForConfig 使用指定客户端配置创建一个代理
func (mgr *defaultProxyManager) NewForConfig(ctx context.Context, config *rest.Config) (*ProxyInfo, error) {
	//TODO implement me
	panic("implement me")
}

// LockConfig 当前进程认领并锁定客户端配置对应的代理（避免其它进程基于此客户端配置启动代理）
func (mgr *defaultProxyManager) LockConfig(ctx context.Context, config *rest.Config) (*ProxyInfo, error) {
	//TODO implement me
	panic("implement me")
}

// UnlockConfig 解锁当前进程认领的客户端配置对应的代理
func (mgr *defaultProxyManager) UnlockConfig(ctx context.Context, config *rest.Config) error {
	//TODO implement me
	panic("implement me")
}

// SetProxyInfoForConfig 设置客户端配置对应的代理信息
// NOTE: 仅能设置当前进程提供的代理服务信息，需要先 LockConfig
func (mgr *defaultProxyManager) SetProxyInfoForConfig(
	ctx context.Context,
	config *rest.Config,
	info *ProxyInfo,
) error {
	//TODO implement me
	panic("implement me")
}

// getConfigSignature 获取客户端配置签名
func (mgr *defaultProxyManager) getConfigSignature(config *rest.Config) string {
	//TODO implement me
	panic("implement me")
}
