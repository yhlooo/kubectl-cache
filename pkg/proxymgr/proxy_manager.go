package proxymgr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-logr/logr"
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
func NewProxyManager(dataRoot string, startProxyArgs []string) ProxyManager {
	return &defaultProxyManager{
		dataRoot:       dataRoot,
		startProxyArgs: startProxyArgs,
	}
}

const (
	rootSubPath     = "kubectl_cache_proxies"
	pidFileSubPath  = "proxy.pid"
	portFileSubPath = "proxy_port"
)

// defaultProxyManager 是 ProxyManager 的一个默认实现
type defaultProxyManager struct {
	dataRoot       string
	startProxyArgs []string
}

var _ ProxyManager = &defaultProxyManager{}

// List 列出所有正在运行的代理
func (mgr *defaultProxyManager) List(ctx context.Context) ([]ProxyInfo, error) {
	//TODO implement me
	panic("implement me")
}

// GetForConfig 获取使用指定客户端配置的代理
func (mgr *defaultProxyManager) GetForConfig(_ context.Context, config *rest.Config) (*ProxyInfo, error) {
	info := &ProxyInfo{}
	info.clientConfigSignature = GetConfigSignature(config)
	info.dataRoot = filepath.Join(mgr.dataRoot, rootSubPath, info.clientConfigSignature)

	// 读 pid 文件
	pidFilePath := filepath.Join(info.dataRoot, pidFileSubPath)
	pidStr, err := os.ReadFile(pidFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("proxy for config does not exist")
		}
		return nil, fmt.Errorf("read proxy pid file %q error: %w", pidFilePath, err)
	}
	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		return nil, fmt.Errorf("invalid proxy pid %q: %w", string(pidStr), err)
	}
	info.pid = pid

	// 读 port 文件
	portFilePath := filepath.Join(info.dataRoot, portFileSubPath)
	portStr, err := os.ReadFile(portFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("proxy for config does not ready")
		}
		return nil, fmt.Errorf("read proxy port file %q error: %w", portFilePath, err)
	}
	port, err := strconv.Atoi(string(portStr))
	if err != nil {
		return nil, fmt.Errorf("invalid proxy port %q: %w", string(pidStr), err)
	}
	info.port = port

	return info, nil
}

// NewForConfig 使用指定客户端配置创建一个代理
func (mgr *defaultProxyManager) NewForConfig(ctx context.Context, config *rest.Config) (*ProxyInfo, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// 启动代理
	cmd := exec.Command(os.Args[0], mgr.startProxyArgs...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start proxy error: %w", err)
	}

	// 等待代理就绪
	lastLogTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for proxy ready error: %w", ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}

		info, err := mgr.GetForConfig(ctx, config)
		if err != nil {
			if time.Since(lastLogTime) >= time.Second {
				logger.V(1).Info(fmt.Sprintf("waiting for proxy ready ... (%s)", err))
				lastLogTime = time.Now()
			}
			continue
		}
		return info, nil
	}
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
