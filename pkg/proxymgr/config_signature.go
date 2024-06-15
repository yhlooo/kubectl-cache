package proxymgr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"k8s.io/client-go/rest"
)

// GetConfigSignature 计算客户端配置签名
func GetConfigSignature(config *rest.Config) string {
	raw, _ := json.Marshal(map[string]interface{}{
		"Host":               config.Host,
		"APIPath":            config.APIPath,
		"Username":           config.Username,
		"Password":           config.Password,
		"BearerToken":        config.BearerToken,
		"BearerTokenFile":    config.BearerTokenFile,
		"Impersonate":        config.Impersonate,
		"AuthProvider":       config.AuthProvider,
		"ExecProvider":       config.ExecProvider,
		"TLSClientConfig":    config.TLSClientConfig,
		"UserAgent":          config.UserAgent,
		"DisableCompression": config.DisableCompression,
		"QPS":                config.QPS,
		"Burst":              config.Burst,
		"Timeout":            config.Timeout,
		"AcceptContentTypes": config.AcceptContentTypes,
		"ContentType":        config.ContentType,
		"GroupVersion":       config.GroupVersion,
	})
	sum := sha256.Sum256(raw)
	// NOTE: 只取前 4 个字节（ 8 个十六进制符号）是因为 UNIX Socket 地址长度不能超过 108 字节（ linux ）或 104 字节（ darwin ）
	return hex.EncodeToString(sum[:4])
}
