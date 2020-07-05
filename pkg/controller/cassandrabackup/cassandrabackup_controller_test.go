package cassandrabackup

import (
	"testing"

	csd "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
	"github.com/stretchr/testify/assert"
)

func TestParseBandwidth(t *testing.T) {
	assert := assert.New(t)

	value, err := parseBandwidth("250")
	assert.Nil(err)
	assert.Equal(value, &csd.DataRate{Value: 250, Unit: "BPS"})

	value, err = parseBandwidth("10k")
	assert.Nil(err)
	assert.Equal(value, &csd.DataRate{Value: 10, Unit: "KBPS"})

	value, err = parseBandwidth("1024M")
	assert.Nil(err)
	assert.Equal(value, &csd.DataRate{Value: 1024, Unit: "MBPS"})

	value, err = parseBandwidth("10G")
	assert.Nil(err)
	assert.Equal(value, &csd.DataRate{Value: 10, Unit: "GBPS"})

	value, err = parseBandwidth("250T")
	assert.NotNil(err)
	assert.Nil(value)

	value, err = parseBandwidth("0.25M")
	assert.NotNil(err)
	assert.Nil(value)
}
