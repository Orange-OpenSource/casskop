package backrest

import (
	"testing"

	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/sidecarclient"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestPerformRestore(t *testing.T) {
	assert := assert.New(t)

	sr := SidecarRestore{
		CoordinatorMember:  "podA",
		csClient: sidecarclient.NewMockCassandraSidecarClient(),
	}

	var concurrentConnection int32 = 15

	cr := &v1alpha1.CassandraRestore{
		Spec:       v1alpha1.CassandraRestoreSpec{
			ConcurrentConnection: &concurrentConnection,
			NoDeleteTruncates: true,
			RestorationStrategyType: "HARDLINKS",
			Cluster: &corev1.LocalObjectReference{
				Name: "cassandra-bgl",
			},
			Backup: &corev1.LocalObjectReference{
				Name: "gcp_backup",
			},
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
		TimeCreated:      "2020-06-10T04:53:05.976Z",
		TimeStarted:      "2020-06-10T05:53:05.976Z",
		TimeCompleted:    "2020-06-10T06:53:05.976Z",
		Condition:        &v1alpha1.RestoreCondition{Type: v1alpha1.RestorePending, LastTransitionTime: cs.Condition.LastTransitionTime},
		Progress:         "10%",
		RestorationPhase: v1alpha1.RestorationPhaseDownload,
		Id:               cs.Id,
	}, cs)

	sr = SidecarRestore{
		CoordinatorMember:  "podA",
		csClient: sidecarclient.NewMockCassandraSidecarClientFailOps(),
	}

	cs, err = sr.PerformRestore(cr, cb)
	assert.Equal(sidecarclient.ErrCassandraSidecarNotReturned201, err)
	assert.Nil(cs)
}

func TestGetRestorebyId(t *testing.T) {
	assert := assert.New(t)

	sr := SidecarRestore{
		CoordinatorMember:  "podA",
		csClient: sidecarclient.NewMockCassandraSidecarClient(),
	}

	operationId := "d3262073-8101-450f-9a11-c851760abd57"
	cs, err := sr.GetRestorebyId(operationId)

	assert.Nil(err)
	assert.NotNil(cs)
	assert.Equal(&v1alpha1.CassandraRestoreStatus{
		TimeCreated:      "2020-06-10T04:53:05.976Z",
		TimeStarted:      "2020-06-10T05:53:05.976Z",
		TimeCompleted:    "2020-06-10T06:53:05.976Z",
		Condition:        &v1alpha1.RestoreCondition{Type: v1alpha1.RestoreRunning, LastTransitionTime: cs.Condition.LastTransitionTime},
		Progress:         "10%",
		RestorationPhase: v1alpha1.RestorationPhaseTruncate,
		Id:               operationId,
	}, cs)

	sr = SidecarRestore{
		CoordinatorMember:  "podA",
		csClient: sidecarclient.NewMockCassandraSidecarClientFailOps(),
	}

	cs, err = sr.GetRestorebyId(operationId)
	assert.Equal(sidecarclient.ErrCassandraSidecarNotReturned200, err)
	assert.Nil(cs)
}