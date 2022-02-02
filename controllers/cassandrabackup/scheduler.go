package cassandrabackup

import (
	"fmt"

	api "github.com/Orange-OpenSource/casskop/api/v2"
	cron "github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

type entryType struct {
	ID       cron.EntryID
	Schedule string
}
type Scheduler struct {
	entries    map[string]entryType
	cronClient *cron.Cron
}

func NewScheduler() Scheduler {
	cronClient := cron.New()
	cronClient.Start()
	return Scheduler{entries: make(map[string]entryType), cronClient: cronClient}
}

// Contains check if a cron task already exists
func (schedule Scheduler) Contains(backupName string) bool {
	_, found := schedule.entries[backupName]
	return found
}

// AddOrUpdate a cron task
func (schedule Scheduler) AddOrUpdate(cassandraBackup *api.CassandraBackup,
	task func(), recorder *record.EventRecorder) (skipped bool, err error) {

	backupName := cassandraBackup.Name

	if schedule.Contains(backupName) && schedule.entries[backupName].Schedule == cassandraBackup.Spec.Schedule {
		return true, nil
	}

	schedule.Remove(backupName)

	entryID, err := schedule.cronClient.AddFunc(cassandraBackup.Spec.Schedule, task)
	if err == nil {
		schedule.entries[backupName] = entryType{entryID, cassandraBackup.Spec.Schedule}
		(*recorder).Event(
			cassandraBackup,
			corev1.EventTypeNormal,
			"BackupTaskscheduled",
			fmt.Sprintf("Controller scheduled task %s to back up cluster %s under snapshot %s with schedule %s",
				cassandraBackup.Name, cassandraBackup.Spec.CassandraCluster, cassandraBackup.Spec.SnapshotTag,
				cassandraBackup.Spec.Schedule))
	}
	return false, err
}

func (schedule Scheduler) Remove(backupName string) {
	if schedule.Contains(backupName) {
		schedule.cronClient.Remove(schedule.entries[backupName].ID)
		delete(schedule.entries, backupName)
	}
}
