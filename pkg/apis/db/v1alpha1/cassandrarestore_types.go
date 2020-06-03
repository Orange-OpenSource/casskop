package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestoreConditionType represents a valid condition of a Restore.
type RestoreConditionType string

const (
	// RestoreScheduled means the Restore has been assigned to a Cluster
	// member for execution.
	RestoreScheduled RestoreConditionType = "Scheduled"
	// RestoreRunning means the Restore is currently being executed by a
	// Cluster member's mysql-agent side-car.
	RestoreRunning RestoreConditionType = "Running"
	// RestoreComplete means the Restore has successfully executed and the
	// resulting artifact has been stored in object storage.
	RestoreComplete RestoreConditionType = "Complete"
	// RestoreFailed means the Restore has failed.
	RestoreFailed RestoreConditionType = "Failed"
)

// RestoreCondition describes the observed state of a Restore at a certain point.
type RestoreCondition struct {
	Type   RestoreConditionType `json:"type"`
	Status corev1.ConditionStatus `json:"status"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

// CassandraRestoreSpec defines the specification for a restore of a Cassandra backup.
type CassandraRestoreSpec struct {
	// Cluster is a refeference to the Cluster to which the Restore
	// belongs.
	Cluster *corev1.LocalObjectReference `json:"cluster"`
	// Backup is a reference to the Backup object to be restored.
	Backup *corev1.LocalObjectReference `json:"backup"`
	// ScheduledMember is the Pod name of the Cluster member on which the
	// Restore will be executed.
	ScheduledMember string `json:"scheduledMember"`
}

// CassandraRestoreStatus captures the current status of a Cassandra restore.
type CassandraRestoreStatus struct {
	// TimeStarted is the time at which the restore was started.
	// +optional
	TimeStarted metav1.Time `json:"timeStarted"`
	// TimeCompleted is the time at which the restore completed.
	// +optional
	TimeCompleted metav1.Time `json:"timeCompleted"`
	// +optional
	Conditions []RestoreCondition `json:"conditions,omitempty"`
}

// +genclient
// +genclient:noStatus
// +resourceName=cassandrarestores
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraRestore is a Casskop Operator resource that represents the restoration of
// backup of a Cassandra cluster.
type CassandraRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CassandraRestoreSpec   `json:"spec"`
	Status CassandraRestoreStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraRestoreList is a list of Restores.
type CassandraRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CassandraRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CassandraRestore{}, &CassandraRestoreList{})
}

// GetRestoreCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetRestoreCondition(status *CassandraRestoreStatus, conditionType RestoreConditionType) (int, *RestoreCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}