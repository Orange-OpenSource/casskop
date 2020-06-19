package cassandrabackup

import (
	"fmt"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
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
	CronClient *cron.Cron
}

// NewScheduler return a new scheduler
func NewScheduler() Scheduler {
	cronClient := cron.New()
	cronClient.Start()
	return Scheduler{entries: make(map[string]entryType), CronClient: cronClient}
}

// Contains check if a cron task already exists
func (schedule Scheduler) Contains(backupName string) bool {
	_, found := schedule.entries[backupName]
	return found
}

// AddOrUpdate add or update a cron task
func (schedule Scheduler) AddOrUpdate(cb *api.CassandraBackup,
	task func(), recorder *record.EventRecorder) (skip bool, err error) {

	backupName := cb.Name

	// Do nothing when backup is already in the cron with the same schedule
	if schedule.Contains(backupName) && schedule.entries[backupName].Schedule == cb.Spec.Schedule {
		return true, nil
	}

	schedule.Remove(backupName)

	entryID, err := schedule.CronClient.AddFunc(cb.Spec.Schedule, task)
	if err == nil {
		schedule.entries[backupName] = entryType{entryID, cb.Spec.Schedule}
		(*recorder).Event(
			cb,
			corev1.EventTypeNormal,
			"BackupTaskscheduled",
			fmt.Sprintf("Controller scheduled task %s to back up cluster %s under snapshot %s with schedule %s",
				cb.Name, cb.Spec.CassandraCluster, cb.Spec.SnapshotTag, cb.Spec.Schedule))

	}
	return false, err
}

// Remove a function from cron client
func (schedule Scheduler) Remove(backupName string) {
	if schedule.Contains(backupName) {
		schedule.CronClient.Remove(schedule.entries[backupName].ID)
		delete(schedule.entries, backupName)
	}
}
