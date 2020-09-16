package cassandrabackup

import (
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
	client.config.UseSSL = true
	client.config.TLSConfig = &tls.Config{}

	conf := client.cassandraBackupSidecarConfig()

	assert.Equal("HTTPS", conf.Scheme)
	assert.Equal(fmt.Sprintf("https://%s:4567", hostnamePodA), conf.BasePath)
}
