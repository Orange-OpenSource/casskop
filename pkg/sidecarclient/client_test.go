package sidecarclient

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNew(t *testing.T) {
	opts := newMockOpts()
	if client := New(opts); client == nil {
		t.Error("Expected new client, got nil")
	}
}

func TestBuild(t *testing.T) {
	client := newMockClient()
	if err := client.Build(); err != nil {
		t.Error("Expected to build mock client, got error:", err)
	}
}

func TestGetCassandraPodSidecarConfig(t *testing.T) {
	assert := assert.New(t)

	client := newMockClient()
	client.opts.UseSSL = true
	client.opts.TLSConfig = &tls.Config{}

	hostname := "podA.test.cassandra.svc.cluster.local"
	podIp := "10.110.105.30"

	conf := client.getCassandraPodSidecarConfig(&v1.Pod{
		Spec: v1.PodSpec{
			Hostname: hostname,
		},
		Status: v1.PodStatus{
			PodIP: podIp,
		},
	})

	assert.Equal("HTTPS", conf.Scheme)
	assert.Equal("https://10.110.105.30:4567", conf.BasePath)
	assert.Equal(hostname, conf.Host)
}
