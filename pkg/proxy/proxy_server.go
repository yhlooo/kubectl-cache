package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/proxy"
	"k8s.io/kubectl/pkg/util"
)

// ServerOptions 代理服务选项
type ServerOptions struct {
	// Kubernetes 客户端配置
	ClientConfig *rest.Config
	// Kubernetes API 映射器
	RESTMapper meta.RESTMapper

	// 监听选项
	Listener ListenerOptions

	// Kubernetes API 代理
	APIProxy APIProxyServerOptions
	// 静态文件服务
	Static StaticServerOptions

	// 最大空闲时间（超过后代理服务自行关闭）
	MaxIdleTime time.Duration
}

// ListenerOptions 监听器选项
type ListenerOptions struct {
	TCP        *TCPListenerOptions
	UNIXSocket *UNIXSocketListenerOptions
}

// UNIXSocketListenerOptions 监听 UNIX Socket 选项
type UNIXSocketListenerOptions struct {
	Path string
}

// TCPListenerOptions 监听 TCP 选项
type TCPListenerOptions struct {
	Address string
	Port    int
}

// APIProxyServerOptions API 代理服务选项
type APIProxyServerOptions struct {
	URIPrefix          string
	Filter             *proxy.FilterServer
	Keepalive          time.Duration
	AppendLocationPath bool
}

// StaticServerOptions 静态服务选项
type StaticServerOptions struct {
	URIPrefix string
	FileBase  string
}

// NewServer 创建一个代理服务
func NewServer(ctx context.Context, opts ServerOptions) (*Server, error) {
	s := &Server{
		readyCh:     make(chan struct{}),
		maxIdleTime: opts.MaxIdleTime,
	}

	handler, err := NewProxyHandler(
		ctx,
		opts.APIProxy.URIPrefix,
		opts.APIProxy.Filter,
		opts.ClientConfig,
		opts.RESTMapper,
		opts.APIProxy.Keepalive,
		opts.APIProxy.AppendLocationPath,
		s.Notify,
	)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle(opts.APIProxy.URIPrefix, handler)
	if opts.Static.FileBase != "" {
		mux.Handle(
			opts.Static.URIPrefix,
			http.StripPrefix(opts.Static.URIPrefix, http.FileServer(http.Dir(opts.Static.FileBase))),
		)
	}
	s.handler = mux

	var listen ListenFunc
	switch {
	case opts.Listener.TCP != nil:
		listen = listenTCP(opts.Listener.TCP.Address, opts.Listener.TCP.Port)
	case opts.Listener.UNIXSocket != nil:
		listen = listenUNIX(opts.Listener.UNIXSocket.Path)
	default:
		return nil, fmt.Errorf("one of opts.Listener.TCP and opts.Listener.UNIXSocket must be specified")
	}
	s.listen = listen

	return s, nil
}

// Server 代理服务
type Server struct {
	server  *http.Server
	handler http.Handler

	readyCh chan struct{}

	listen   ListenFunc
	listener net.Listener

	idleTimerLock sync.Mutex
	idleTimer     *time.Timer
	maxIdleTime   time.Duration
}

// Serve 开始提供 HTTP 服务
func (s *Server) Serve(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	// 开始监听
	var err error
	s.listener, err = s.listen()
	if err != nil {
		close(s.readyCh)
		return err
	}
	logger.V(1).Info(fmt.Sprintf("serve on %q", s.listener.Addr()))

	// 创建 HTTP Server
	s.server = &http.Server{
		Handler: s.handler,
	}

	// 等待 ctx 结束
	var ctxErr error
	go func() {
		<-ctx.Done()
		ctxErr = ctx.Err()
		logger.V(1).Info("context done, shutting down server ...")
		if err := s.server.Shutdown(logr.NewContext(context.Background(), logger)); err != nil {
			logger.Error(err, "shutdown error")
		}
	}()

	// 设置空闲停止倒计时
	if s.maxIdleTime > 0 {
		s.idleTimer = time.NewTimer(s.maxIdleTime)
		go func() {
			select {
			case <-ctx.Done():
			case <-s.idleTimer.C:
				logger.Info("idle timeout, shutting down server ...")
				if err := s.server.Shutdown(ctx); err != nil {
					logger.Error(err, "shutdown error")
				}
			}
		}()
	}

	// 开始 HTTP 服务
	close(s.readyCh)
	logger.Info(fmt.Sprintf("Starting to serve on %s", s.listener.Addr()))
	serveErr := s.server.Serve(s.listener)
	if ctxErr != nil {
		return ctxErr
	}
	return serveErr
}

// Stop 停止服务
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// Ready 返回一个通道，该通道在服务端就绪时会被关闭
func (s *Server) Ready() <-chan struct{} {
	return s.readyCh
}

// Addr 返回监听地址
func (s *Server) Addr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Notify 通知收到了一个请求
func (s *Server) Notify(_ *http.Request) {
	if s.maxIdleTime == 0 {
		return
	}

	// 重置计时器
	s.idleTimerLock.Lock()
	defer s.idleTimerLock.Unlock()
	s.idleTimer.Reset(s.maxIdleTime)
}

// ListenFunc 开始监听方法
type ListenFunc func() (net.Listener, error)

// Listen 监听 TCP
func listenTCP(address string, port int) ListenFunc {
	return func() (net.Listener, error) {
		return net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	}
}

// listenUNIX 监听 UNIX Socket
func listenUNIX(path string) ListenFunc {
	return func() (net.Listener, error) {
		// Remove any socket, stale or not, but fall through for other files
		fi, err := os.Stat(path)
		if err == nil && (fi.Mode()&os.ModeSocket) != 0 {
			_ = os.Remove(path)
		}
		// Default to only user accessible socket, caller can open up later if desired
		oldmask, _ := util.Umask(0077)
		l, err := net.Listen("unix", path)
		_, _ = util.Umask(oldmask)
		return l, err
	}
}
