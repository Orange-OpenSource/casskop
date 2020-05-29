package cassandrabackup

import (
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	cron "github.com/robfig/cron/v3"
)

type Scheduler struct {
	entries    map[string]cron.EntryID
	CronClient *cron.Cron
}

// NewScheduler return a new scheduler
func NewScheduler() Scheduler {
	return Scheduler{entries: make(map[string]cron.EntryID), CronClient: cron.New()}
}

// Contains check if a cron task already exists
func (schedule Scheduler) Contains(backupName string) bool {
	_, found := schedule.entries[backupName]
	return found
}

// AddOrUpdate add or update a cron task
func (schedule Scheduler) AddOrUpdate(backupName string, backupSpec api.CassandraBackupSpec,
	task func()) (err error) {

	schedule.Remove(backupName)

	entryID, err := schedule.CronClient.AddFunc(backupSpec.Schedule, task)
	if err != nil {
		schedule.entries[backupName] = entryID
	}
	return
}

// Remove a function from cron client
func (schedule Scheduler) Remove(backupName string) {
	if schedule.Contains(backupName) {
		schedule.CronClient.Remove(schedule.entries[backupName])
		delete(schedule.entries, backupName)
	}
}
