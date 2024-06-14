package proxymgr

// ProxyInfo 代理信息
type ProxyInfo struct {
	pid  int
	port int

	clientConfigSignature string
	dataRoot              string
}

// PID 代理服务进程 ID
func (info *ProxyInfo) PID() int {
	return info.pid
}

// TCPPort 代理服务 TCP 端口
func (info *ProxyInfo) TCPPort() int {
	return info.port
}

// SetTCPPort 设置代理 TCP 端口
func (info *ProxyInfo) SetTCPPort(port int) {
	info.port = port
}

// DataRoot 代理服务数据存储根路径
func (info *ProxyInfo) DataRoot() string {
	return info.dataRoot
}

// ClientConfigSignature 代理服务的客户端配置签名
func (info *ProxyInfo) ClientConfigSignature() string {
	return info.clientConfigSignature
}
