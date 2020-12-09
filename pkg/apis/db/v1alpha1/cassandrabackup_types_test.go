// Copyright 2019 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
)

func TestValidateSchedule(t *testing.T) {
	assert := assert.New(t)

	backupSpec := &CassandraBackupSpec{
		Schedule: "",
	}

	for _, schedule := range []string{"@midnight", "1 1 */2 11 *", "@daily"} {
		backupSpec.Schedule = schedule
		assert.Nilf(backupSpec.ValidateScheduleFormat(), "Schedule %s should be parseable", schedule)
	}

	backupSpec.Schedule = "@noon" // Unknown descriptor
	assert.NotNilf(backupSpec.ValidateScheduleFormat(), "Schedule %s should not be parseable", backupSpec.Schedule)
}

func TestCassandraBackupComputeLastAppliedConfiguration(t *testing.T) {
	assert := assert.New(t)

	backupSpec := &CassandraBackupSpec{
		CassandraCluster: "cluster1",
		Datacenter:       "dc1",
		Schedule:         "@weekly",
		StorageLocation:  "s3://cassie",
		SnapshotTag:      "weekly",
		Entities:         "k1.t1, k3.t3",
	}

	backup := &CassandraBackup{
		Spec: *backupSpec,
	}

	lastAppliedConfiguration, _ := backup.ComputeLastAppliedAnnotation()
	result := `{"metadata":{"creationTimestamp":null},
                "spec":{"cassandracluster":"cluster1","datacenter":"dc1","storageLocation":"s3://cassie",
                        "schedule":"@weekly","snapshotTag":"weekly","entities":"k1.t1, k3.t3"},"status":{}
                }`

	comparison, _ := jsondiff.Compare([]byte(lastAppliedConfiguration), []byte(result), &jsondiff.Options{})

	assert.Equal(jsondiff.FullMatch, comparison)
}
