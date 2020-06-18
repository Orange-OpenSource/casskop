package v1alpha1

import (
	"fmt"
	"strconv"

	"github.com/Orange-OpenSource/casskop/pkg/util"
	csapi "github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
	"github.com/mitchellh/mapstructure"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestoreConditionType represents a valid condition of a Restore.
type RestoreConditionType string

const (
	// RestoreRequired means the Restore has been assigned to a Cluster
	// member for execution.
	RestoreRequired RestoreConditionType = "REQUIRED"
	// RestorePending means the Restore operation is pending for being submitted.
	RestorePending RestoreConditionType = "PENDING"
	// RestoreRunning means the Restore is currently being executed by a
	// Cluster member's mysql-agent side-car.
	RestoreRunning RestoreConditionType = "RUNNING"
	// RestoreComplete means the Restore has successfully executed and the
	// resulting artifact has been stored in object storage.
	RestoreCompleted RestoreConditionType = "COMPLETED"
	// RestoreFailed means the Restore has failed.
	RestoreFailed RestoreConditionType = "FAILED"
	// RestoreCanceled means the Restore operation was interrupted while being run.
	RestoreCanceled RestoreConditionType = "CANCELED"
)

func (r RestoreConditionType) IsInProgress() bool {
	return r == RestorePending || r == RestoreRunning
}

func (r RestoreConditionType) IsInError() bool {
	return r == RestoreFailed || r == RestoreCanceled
}

func (r RestoreConditionType) IsRequired() bool {
	return r == RestoreRequired
}

func (r RestoreConditionType) IsPending() bool {
	return r == RestorePending
}

func (r RestoreConditionType) IsRunning() bool {
	return r == RestoreRunning
}

func (r RestoreConditionType) IsCompleted() bool {
	return r == RestoreCompleted
}


// RestoreCondition describes the observed state of a Restore at a certain point.
type RestoreCondition struct {
	Type   RestoreConditionType `json:"type"`
	// +optional
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// +optional
	FailureCause *FailureCause `json:"failureCause,omitempty"`
}

type FailureCause struct {
	StackTrace       []StackTrace  `json:"stackTrace"`
	Message          string        `json:"message"`
	LocalizedMessage string        `json:"localizedMessage"`
}

type StackTrace struct {
	MethodName   string `json:"methodName,omitempty"`
	FileName     string `json:"fileName,omitempty"`
	LineNumber   string `json:"lineNumber,omitempty"`
	ClassName    string `json:"className,omitempty"`
	NativeMethod string `json:"nativeMethod,omitempty"`
}

type RestorationPhaseType string

const (
	RestorationPhaseDownload RestorationPhaseType = "DOWNLOAD"
	RestorationPhaseImport RestorationPhaseType = "IMPORT"
	RestorationPhaseTruncate RestorationPhaseType = "TRUNCATE"
	RestorationPhaseCleanup RestorationPhaseType = "CLEANUP"
	RestorePhaseUnknown RestorationPhaseType = "UNKNOWN"
)


// CassandraRestoreStatus captures the current status of a Cassandra restore.
type CassandraRestoreStatus struct {
	//
	TimeCreated string `json:"timeCreated,omitempty"`
	// TimeStarted is the time at which the restore was started.
	TimeStarted string `json:"timeStarted,omitempty"`
	// TimeCompleted is the time at which the restore completed.
	TimeCompleted string `json:"timeCompleted,omitempty"`
	//
	Condition *RestoreCondition `json:"conditions,omitempty"`
	// Progress is a float from 0.0 to 1.0, 1.0 telling that operation is completed, either successfully or with errors
	Progress string `json:"progress,omitempty"`
	//
	RestorationPhase RestorationPhaseType `json:"restorationPhase,omitempty"`
	// unique identifier of an operation, a random id is assigned to each operation after a request is submitted,
	// from caller's perspective, an id is sent back as a response to his request so he can further query state of that operation,
	// referencing id, by operations/{id} endpoint
	Id string `json:"id,omitempty"`
}

// CassandraRestoreSpec defines the specification for a restore of a Cassandra backup.
type CassandraRestoreSpec struct {
	// Cluster is a refeference to the Cluster to which the Restore
	// belongs.
	Cluster *corev1.LocalObjectReference `json:"cluster"`
	// Backup is a reference to the Backup object to be restored.
	Backup *corev1.LocalObjectReference `json:"backup"`
	// CoordinatorMember is the Pod name of the Cluster member on which the
	// Restore will be executed.
	CoordinatorMember string `json:"coordinatorMember,omitempty"`
	// ConcurrentConnection is the number of threads used for upload, there might be
	// at most so many uploading threads at any given time, when not set, it defaults to 10
	ConcurrentConnection *int32 `json:"concurrentConnection,omitempty"`
	// directory of Cassandra, by default it is /var/lib/cassandra, in this path, one expects there is 'data' directory
	CassandraDirectory string `json:"cassandraDirectory,omitempty"`
	// NoDeleteTruncates is flag saying if we should not delete truncated SSTables
	// after they are imported, as part of CLEANUP phase, defaults to false
	NoDeleteTruncates bool `json:"noDeleteTruncates,omitempty"`
	// SchemaVersion is the version of schema we want to restore from, upon backup, a schema version is automatically appended to snapshot name
	// and its manifest is uploaded under that name. In case we have two snapshots having same name,
	// we might distinguish between them by this schema version. If schema version is not specified,
	// we expect that there will be one and only one backup taken with respective snapshot name.
	// This schema version has to match the version of a Cassandra node we are doing restore for
	// (hence, by proxy, when global request mode is used, all nodes have to be on exact same schema version).
	SchemaVersion string `json:"schemaVersion,omitempty"`
	// ExactSchemaVersion is a flag saying if we indeed want a schema version of a running node match with schema version a snapshot is taken on.
	// There might be cases when we want to restore a table for which its CQL schema has not changed
	// but it has changed for other table / keyspace but a schema for that node has changed by doing that.
	ExactSchemaVersion bool `json:"exactSchemaVersion,omitempty""`
	// strategy telling how we should go about restoration, please refer to details in backup and sidecar documentation
	// +kubebuilder:validation:Enum={"HARDLINKS","IMPORT"}
	RestorationStrategyType string `json:"restorationStrategyType,omitempty"`
	// name of Kubernetes secret from which credentials used for the communication to cloud storage providers are read,
	// if not specified, secret name to be read will be automatically derived in form 'cassandra-backup-restore-secret-cluster-{name-of-cluster}'.
	// These secrets are used only in case protocol in storageLocation is gcp, azure or s3.
	SecretName string `json:"secret,omitempty"`
	// database entities to restore, it might be either only keyspaces or only tables (from different keyspaces if needed),
	// e.g. 'k1,k2' if one wants to backup whole keyspaces and 'ks1.t1,ks2,t2' if one wants to restore tables.
	// These formats can not be used together so 'k1,k2.t2' is invalid. If this field is empty, all keyspaces are restored.
	Entities string `json:"entities,omitempty"`
}

// +genclient0
// +genclient:noStatus
// +resourceName=cassandrarestores
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraRestore is a Casskop Operator resource that represents the restoration of
// backup of a Cassandra cluster.
type CassandraRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CassandraRestoreSpec   `json:"spec"`
	Status CassandraRestoreStatus `json:"status,omitempty"`
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
func GetRestoreCondition(status *CassandraRestoreStatus, conditionType RestoreConditionType) *RestoreCondition {
	if status.Condition != nil && status.Condition.Type == conditionType {
		return status.Condition
	}
	return nil
}

func ComputeStatusFromRestoreOperation(restore *csapi.RestoreOperationResponse) CassandraRestoreStatus{
	status := CassandraRestoreStatus{
		Progress:      fmt.Sprintf("%v%%", strconv.Itoa(int(restore.Progress*100))),
		Id:            restore.Id,
		TimeCreated:   restore.CreationTime,
		TimeStarted:   restore.StartTime,
		TimeCompleted: restore.CompletionTime,
		Condition: &RestoreCondition{},
	}

	status.setRestorationPhaseFromString(restore.RestorationPhase)
	status.setRestorationConditionFromRestoreOperation(restore)

	return status
}

func (status *CassandraRestoreStatus) setRestorationPhaseFromString(phase string) {
	switch phase {
	case string(RestorationPhaseDownload):
		status.RestorationPhase = RestorationPhaseDownload
	case string(RestorationPhaseImport):
		status.RestorationPhase = RestorationPhaseImport
	case string(RestorationPhaseTruncate):
		status.RestorationPhase = RestorationPhaseTruncate
	case string(RestorationPhaseCleanup):
		status.RestorationPhase = RestorationPhaseCleanup
	default:
		status.RestorationPhase = RestorePhaseUnknown
	}
}

func (status *CassandraRestoreStatus) setRestorationConditionFromRestoreOperation(restore *csapi.RestoreOperationResponse) {
	status.Condition.LastTransitionTime = metav1.Now().Format(util.TimeStampLayout)
	status.setRestorationStateFromString(restore.State)

	if status.Condition.Type.IsInError() {
		var failureCause FailureCause
		mapstructure.Decode(*restore.FailureCause, &failureCause)

		status.Condition.FailureCause = &failureCause
	}
}

func (status *CassandraRestoreStatus) setRestorationStateFromString(state string) {
	switch state {
	case string(RestoreRequired):
		status.Condition.Type = RestoreRequired
	case string(RestorePending):
		status.Condition.Type = RestorePending
	case string(RestoreRunning):
		status.Condition.Type = RestoreRunning
	case string(RestoreCompleted):
		status.Condition.Type = RestoreCompleted
	case string(RestoreFailed):
		status.Condition.Type = RestoreFailed
	case string(RestoreCanceled):
		status.Condition.Type = RestoreCanceled
	}
}