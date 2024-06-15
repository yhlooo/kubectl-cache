package proxymgr

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

// Proxy 缓存代理
type Proxy struct {
	metav1.TypeMeta `json:",inline"`
	// 对象元信息
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// 代理状态
	Status ProxyStatus `json:"status,omitempty"`
}

var _ runtime.Object = &Proxy{}

// ProxyStatus 代理状态
type ProxyStatus struct {
	// 状态
	State ProxyState `json:"state,omitempty"`
	// 处于当前状态的原因，可枚举
	Reason string `json:"reason,omitempty"`
	// 关于当前状态的人类可读的描述
	Message string `json:"message,omitempty"`
	// 代理服务进程 ID
	PID int `json:"pid,omitempty"`
	// 代理服务监听端口
	Port int `json:"port,omitempty"`
	// 代理服务数据根目录
	DataRoot string `json:"dataRoot,omitempty"`
	// 运行代理的客户端配置签名
	ClientConfigSignature string `json:"clientConfigSignature,omitempty"`
}

// ProxyState 代理状态
type ProxyState string

// ProxyState 的可选值
const (
	ProxyPending ProxyState = "Pending"
	ProxyReady   ProxyState = "Ready"
	ProxyDead    ProxyState = "Dead"
)

// ToClientConfig 返回使用该代理的客户端配置
func (proxy *Proxy) ToClientConfig() *rest.Config {
	return &rest.Config{
		Host: fmt.Sprintf("http://127.0.0.1:%d", proxy.Status.Port),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
}

// ProxyList 缓存代理列表
type ProxyList struct {
	metav1.TypeMeta `json:",inline"`
	// 对象元信息
	metav1.ListMeta `json:"metadata,omitempty"`
	// 代理信息的列表
	Items []Proxy `json:"items"`
}

var _ runtime.Object = (*ProxyList)(nil)

// DeepCopyObject 深拷贝
func (proxy *Proxy) DeepCopyObject() runtime.Object {
	ret := &Proxy{
		TypeMeta: proxy.TypeMeta,
		Status:   proxy.Status,
	}
	proxy.ObjectMeta.DeepCopyInto(&ret.ObjectMeta)
	return ret
}

// DeepCopyObject 深拷贝
func (list *ProxyList) DeepCopyObject() runtime.Object {
	ret := &ProxyList{
		TypeMeta: list.TypeMeta,
	}
	list.ListMeta.DeepCopyInto(&ret.ListMeta)
	if list.Items == nil {
		return ret
	}
	for _, item := range list.Items {
		proxy := item.DeepCopyObject().(*Proxy)
		list.Items = append(list.Items, *proxy)
	}
	return ret
}
