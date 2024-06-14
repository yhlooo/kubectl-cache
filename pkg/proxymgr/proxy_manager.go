package proxymgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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

	// LockProxyInfo 当前进程认领并锁定客户端配置对应的代理（避免其它进程基于此客户端配置启动代理）
	LockProxyInfo(ctx context.Context, config *rest.Config) (*ProxyInfo, error)
	// UnlockProxyInfo 解锁当前进程认领的客户端配置对应的代理
	UnlockProxyInfo(ctx context.Context, info *ProxyInfo) error
	// SetProxyInfo 设置客户端配置对应的代理信息
	// NOTE: 仅能设置当前进程提供的代理服务信息，需要先 LockConfig
	SetProxyInfo(ctx context.Context, info *ProxyInfo) error
}

// NewProxyManager 创建一个代理服务管理器
func NewProxyManager(dataRoot string) ProxyManager {
	return &defaultProxyManager{dataRoot: dataRoot}
}

const (
	rootSubPath     = "kubectl_cache_proxies"
	pidFileSubPath  = "proxy.pid"
	portFileSubPath = "proxy_port"
)

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

// LockProxyInfo 当前进程认领并锁定客户端配置对应的代理（避免其它进程基于此客户端配置启动代理）
func (mgr *defaultProxyManager) LockProxyInfo(_ context.Context, config *rest.Config) (*ProxyInfo, error) {
	info := &ProxyInfo{
		pid:                   os.Getpid(),
		clientConfigSignature: GetConfigSignature(config),
	}

	// 创建数据目录
	info.dataRoot = filepath.Join(mgr.dataRoot, rootSubPath, info.clientConfigSignature)
	if err := os.MkdirAll(info.dataRoot, 0700); err != nil {
		return nil, fmt.Errorf("make directory %q for proxy error: %w", info.dataRoot, err)
	}

	// 打开 pid 文件
	pidFilePath := filepath.Join(info.dataRoot, pidFileSubPath)
	pidFile, err := os.OpenFile(pidFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("open pid file %q for proxy error: %w", pidFilePath, err)
	}
	defer func() {
		_ = pidFile.Close()
	}()
	// 锁 pid 文件
	if err := lockFile(pidFile); err != nil {
		return nil, fmt.Errorf("lock pid file %q for proxy error: %w", pidFilePath, err)
	}
	// 写 pid
	if _, err := pidFile.WriteString(strconv.Itoa(os.Getpid())); err != nil {
		return nil, fmt.Errorf("write pid file %q for proxy error: %w", pidFilePath, err)
	}

	return info, nil
}

// UnlockProxyInfo 解锁当前进程认领的客户端配置对应的代理
func (mgr *defaultProxyManager) UnlockProxyInfo(_ context.Context, info *ProxyInfo) error {
	// 打开 pid 文件
	pidFilePath := filepath.Join(info.dataRoot, pidFileSubPath)
	pidFile, err := os.OpenFile(pidFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open pid file %q for proxy error: %w", pidFilePath, err)
	}
	// 解锁 pid 文件
	if err := unlockFile(pidFile); err != nil {
		_ = pidFile.Close()
		return fmt.Errorf("unlock pid file %q for proxy error: %w", pidFilePath, err)
	}
	_ = pidFile.Close()
	// 删除所有数据文件
	if err := os.RemoveAll(info.dataRoot); err != nil {
		return fmt.Errorf("remove proxy data directory %q error: %w", info.dataRoot, err)
	}
	return nil
}

// SetProxyInfo 设置客户端配置对应的代理信息
// NOTE: 仅能设置当前进程提供的代理服务信息，需要先 LockConfig
func (mgr *defaultProxyManager) SetProxyInfo(_ context.Context, info *ProxyInfo) error {
	// 写端口文件
	if info.port != 0 {
		portFilePath := filepath.Join(info.dataRoot, portFileSubPath)
		portFile, err := os.OpenFile(portFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("open port file %q for proxy error: %w", portFilePath, err)
		}
		defer func() {
			_ = portFile.Close()
		}()
		if _, err := portFile.WriteString(strconv.Itoa(info.port)); err != nil {
			return fmt.Errorf("write port file %q for proxy error: %w", portFilePath, err)
		}
	}
	return nil
}
