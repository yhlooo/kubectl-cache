package proxymgr

import (
	"testing"

	"k8s.io/client-go/rest"
)

// TestGetConfigSignature 测试 getConfigSignature 方法
func TestGetConfigSignature(t *testing.T) {
	ret := GetConfigSignature(&rest.Config{
		Host: "https://1.2.3.4",
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   []byte("test"),
			CertData: []byte("test"),
			KeyData:  []byte("test"),
		},
		BearerToken: "testtoken",
	})
	expected := "64b4a14ceeb9c31a4c3504358fb6e888d15e999bf23d77910f2a19822ded7f4b"
	if ret != expected {
		t.Errorf("expected: %s, got: %s", expected, ret)
	}
}
