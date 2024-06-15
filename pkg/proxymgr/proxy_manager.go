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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// ProxyManager 代理服务管理器
type ProxyManager interface {
	// List 列出所有正在运行的代理
	List(ctx context.Context) (*ProxyList, error)
	// Get 获取指定代理信息
	Get(ctx context.Context, name string) (*Proxy, error)
	// GetForConfig 获取使用指定客户端配置的代理
	GetForConfig(ctx context.Context, config *rest.Config) (*Proxy, error)
	// NewForConfig 使用指定客户端配置创建一个代理
	NewForConfig(ctx context.Context, config *rest.Config) (*Proxy, error)

	// LockProxy 当前进程认领并锁定客户端配置对应的代理（避免其它进程基于此客户端配置启动代理）
	LockProxy(ctx context.Context, config *rest.Config) (*Proxy, error)
	// UnlockProxy 解锁当前进程认领的客户端配置对应的代理
	UnlockProxy(ctx context.Context, proxy *Proxy) error
	// SetProxy 设置客户端配置对应的代理信息
	// NOTE: 仅能设置当前进程提供的代理服务信息，需要先 LockConfig
	SetProxy(ctx context.Context, proxy *Proxy) error
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
func (mgr *defaultProxyManager) List(ctx context.Context) (*ProxyList, error) {
	logger := logr.FromContextOrDiscard(ctx)

	proxiesDirPath := filepath.Join(mgr.dataRoot, rootSubPath)
	proxyDirs, err := os.ReadDir(proxiesDirPath)
	if err != nil {
		return nil, fmt.Errorf("list proxy directories in %q error: %w", proxiesDirPath, err)
	}

	ret := &ProxyList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
	}
	for _, dir := range proxyDirs {
		if !dir.IsDir() {
			continue
		}

		proxy, err := mgr.Get(ctx, dir.Name())
		if err != nil {
			logger.Info(fmt.Sprintf("WARNING get proxy %q error: %v", dir.Name(), err))
			continue
		}
		ret.Items = append(ret.Items, *proxy)
	}

	return ret, nil
}

// Get 获取指定代理信息
func (mgr *defaultProxyManager) Get(_ context.Context, name string) (*Proxy, error) {
	proxy := &Proxy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Proxy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: ProxyStatus{
			DataRoot:              filepath.Join(mgr.dataRoot, rootSubPath, name),
			ClientConfigSignature: name,
		},
	}

	// 读 pid 文件
	var err error
	proxy.Status.PID, err = mgr.getPID(proxy.Status.DataRoot)
	if err != nil {
		return nil, err
	}
	// 检查进程状态
	if _, err := os.FindProcess(proxy.Status.PID); err != nil {
		proxy.Status.State = ProxyDead
		proxy.Status.Reason = "GetProcessError"
		proxy.Status.Message = fmt.Sprintf("get proxy process error: %v", err)
		return proxy, nil
	}

	// 获取端口
	proxy.Status.Port, err = mgr.getPort(proxy.Status.DataRoot)
	if err != nil {
		proxy.Status.State = ProxyPending
		proxy.Status.Reason = "GetProxyPortError"
		proxy.Status.Message = fmt.Sprintf("get proxy port error: %v", err)
		return proxy, nil
	}

	proxy.Status.State = ProxyReady

	return proxy, nil
}

// GetForConfig 获取使用指定客户端配置的代理
func (mgr *defaultProxyManager) GetForConfig(ctx context.Context, config *rest.Config) (*Proxy, error) {
	return mgr.Get(ctx, GetConfigSignature(config))
}

// NewForConfig 使用指定客户端配置创建一个代理
func (mgr *defaultProxyManager) NewForConfig(ctx context.Context, config *rest.Config) (*Proxy, error) {
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

		proxy, err := mgr.GetForConfig(ctx, config)
		if err != nil {
			if time.Since(lastLogTime) >= time.Second {
				logger.V(1).Info(fmt.Sprintf("waiting for proxy ready ... (%s)", err))
				lastLogTime = time.Now()
			}
			continue
		}
		if proxy.Status.State != ProxyReady {
			if time.Since(lastLogTime) >= time.Second {
				logger.V(1).Info(fmt.Sprintf(
					"waiting for proxy ready ... (state: %s, reason: %s, message: %s)",
					proxy.Status.State, proxy.Status.Reason, proxy.Status.Message,
				))
				lastLogTime = time.Now()
			}
			continue
		}
		return proxy, nil
	}
}

// LockProxy 当前进程认领并锁定客户端配置对应的代理（避免其它进程基于此客户端配置启动代理）
func (mgr *defaultProxyManager) LockProxy(_ context.Context, config *rest.Config) (*Proxy, error) {
	proxy := &Proxy{
		Status: ProxyStatus{
			State:                 "",
			PID:                   os.Getpid(),
			Port:                  0,
			DataRoot:              "",
			ClientConfigSignature: GetConfigSignature(config),
		},
	}

	// 创建数据目录
	proxy.Status.DataRoot = filepath.Join(mgr.dataRoot, rootSubPath, proxy.Status.ClientConfigSignature)
	if err := os.MkdirAll(proxy.Status.DataRoot, 0700); err != nil {
		return nil, fmt.Errorf("make directory %q for proxy error: %w", proxy.Status.DataRoot, err)
	}

	// 打开 pid 文件
	pidFilePath := filepath.Join(proxy.Status.DataRoot, pidFileSubPath)
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

	return proxy, nil
}

// UnlockProxy 解锁当前进程认领的客户端配置对应的代理
func (mgr *defaultProxyManager) UnlockProxy(_ context.Context, proxy *Proxy) error {
	// 打开 pid 文件
	pidFilePath := filepath.Join(proxy.Status.DataRoot, pidFileSubPath)
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
	if err := os.RemoveAll(proxy.Status.DataRoot); err != nil {
		return fmt.Errorf("remove proxy data directory %q error: %w", proxy.Status.DataRoot, err)
	}
	return nil
}

// SetProxy 设置客户端配置对应的代理信息
// NOTE: 仅能设置当前进程提供的代理服务信息，需要先 LockConfig
func (mgr *defaultProxyManager) SetProxy(_ context.Context, proxy *Proxy) error {
	// 写端口文件
	if proxy.Status.Port != 0 {
		portFilePath := filepath.Join(proxy.Status.DataRoot, portFileSubPath)
		portFile, err := os.OpenFile(portFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("open port file %q for proxy error: %w", portFilePath, err)
		}
		defer func() {
			_ = portFile.Close()
		}()
		if _, err := portFile.WriteString(strconv.Itoa(proxy.Status.Port)); err != nil {
			return fmt.Errorf("write port file %q for proxy error: %w", portFilePath, err)
		}
	}
	return nil
}

// getPID 获取代理服务进程 ID
func (mgr *defaultProxyManager) getPID(proxyDataRoot string) (int, error) {
	pidFilePath := filepath.Join(proxyDataRoot, pidFileSubPath)
	pidStr, err := os.ReadFile(pidFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("proxy pid file does not exist")
		}
		return 0, fmt.Errorf("read proxy pid file %q error: %w", pidFilePath, err)
	}
	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		return 0, fmt.Errorf("invalid proxy pid %q: %w", string(pidStr), err)
	}
	return pid, nil
}

// getPort 获取代理服务监听端口
func (mgr *defaultProxyManager) getPort(proxyDataRoot string) (int, error) {
	portFilePath := filepath.Join(proxyDataRoot, portFileSubPath)
	portStr, err := os.ReadFile(portFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("proxy does not ready")
		}
		return 0, fmt.Errorf("read proxy port file %q error: %w", portFilePath, err)
	}
	port, err := strconv.Atoi(string(portStr))
	if err != nil {
		return 0, fmt.Errorf("invalid proxy port %q: %w", string(portStr), err)
	}
	return port, nil
}
