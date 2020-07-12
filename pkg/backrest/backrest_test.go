package backrest

import (
	"testing"

	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/cassandrabackup"
	csapi "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
	"github.com/stretchr/testify/assert"
)

func TestPerformRestore(t *testing.T) {
	assert := assert.New(t)

	test := cassandrabackup.NewMockCassandraBackupClient()
	sr := Client{
		CoordinatorMember: "podA",
		client:            test,
	}

	var concurrentConnection int32 = 15

	cr := &v1alpha1.CassandraRestore{
		Spec:       v1alpha1.CassandraRestoreSpec{
			ConcurrentConnection:    &concurrentConnection,
			NoDeleteTruncates:       true,
			RestorationStrategyType: "HARDLINKS",
			CassandraCluster:        "cassandra-bgl",
			CassandraBackup:         "gcp_backup",
		},
	}

	cb := &v1alpha1.CassandraBackup{
		Spec:       v1alpha1.CassandraBackupSpec{
			CassandraCluster: "cassandra-bgl",
			StorageLocation:  "gcp://backup-casskop-aguitton/cassandra-bgl/dc1/cassandra-bgl-dc1-rack1-0",
			SnapshotTag:      "SnapshotTag1",
			Secret:           "cloud-backup-secrets",
			Entities:         "ks1,ks2",
		},
	}

	cs, err := sr.PerformRestore(cr, cb)

	assert.Nil(err)
	assert.NotNil(cs)
	assert.Equal(&v1alpha1.CassandraRestoreStatus{
		TimeCreated:   "2020-06-10T04:53:05.976Z",
		TimeStarted:   "2020-06-10T05:53:05.976Z",
		TimeCompleted: "2020-06-10T06:53:05.976Z",
		Condition:     &v1alpha1.RestoreCondition{Type: v1alpha1.RestorePending, LastTransitionTime: cs.Condition.LastTransitionTime},
		Progress:      "10%",
		Phase:         v1alpha1.RestorationPhaseDownload,
		Id:            cs.Id,
	}, cs)

	sr = Client{
		CoordinatorMember: "podA",
		client:            cassandrabackup.NewMockCassandraBackupClientFailOps(),
	}

	cs, err = sr.PerformRestore(cr, cb)
	assert.Equal(cassandrabackup.ErrCassandraSidecarNotReturned201, err)
	assert.Nil(cs)
}

func TestGetRestorebyId(t *testing.T) {
	assert := assert.New(t)

	c := Client{
		CoordinatorMember: "podA",
		client:            cassandrabackup.NewMockCassandraBackupClient(),
	}

	operationId := "d3262073-8101-450f-9a11-c851760abd57"
	cs, err := c.GetRestoreStatusById(operationId)

	assert.Nil(err)
	assert.NotNil(cs)
	assert.Equal(&v1alpha1.CassandraRestoreStatus{
		TimeCreated:   "2020-06-10T04:53:05.976Z",
		TimeStarted:   "2020-06-10T05:53:05.976Z",
		TimeCompleted: "2020-06-10T06:53:05.976Z",
		Condition:     &v1alpha1.RestoreCondition{Type: v1alpha1.RestoreRunning, LastTransitionTime: cs.Condition.LastTransitionTime},
		Progress:      "10%",
		Phase:         v1alpha1.RestorationPhaseTruncate,
		Id:            operationId,
	}, cs)

	c = Client{
		CoordinatorMember: "podA",
		client:            cassandrabackup.NewMockCassandraBackupClientFailOps(),
	}

	cs, err = c.GetRestoreStatusById(operationId)
	assert.Equal(cassandrabackup.ErrCassandraSidecarNotReturned200, err)
	assert.Nil(cs)
}


func TestParseBandwidth(t *testing.T) {
	assert := assert.New(t)

	value, err := parseBandwidth("250")
	assert.Nil(err)
	assert.Equal(value, &csapi.DataRate{Value: 250, Unit: "BPS"})

	value, err = parseBandwidth("10k")
	assert.Nil(err)
	assert.Equal(value, &csapi.DataRate{Value: 10, Unit: "KBPS"})

	value, err = parseBandwidth("1024M")
	assert.Nil(err)
	assert.Equal(value, &csapi.DataRate{Value: 1024, Unit: "MBPS"})

	value, err = parseBandwidth("10G")
	assert.Nil(err)
	assert.Equal(value, &csapi.DataRate{Value: 10, Unit: "GBPS"})

	value, err = parseBandwidth("250T")
	assert.NotNil(err)
	assert.Nil(value)

	value, err = parseBandwidth("0.25M")
	assert.NotNil(err)
	assert.Nil(value)
}