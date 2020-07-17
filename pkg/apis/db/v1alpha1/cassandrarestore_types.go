package v1alpha1

import (
	"github.com/Orange-OpenSource/casskop/pkg/util"
	csapi "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
	"github.com/mitchellh/mapstructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestoreConditionType represents a valid condition of a Restore
type RestoreConditionType string

const (
	// RestoreRequired means the Restore has been assigned to a Cluster member for execution
	RestoreRequired RestoreConditionType = "REQUIRED"
	RestorePending RestoreConditionType = "PENDING"
	RestoreRunning RestoreConditionType = "RUNNING"
	// RestoreComplete means the Restore has successfully been executed and resulting artifact stored in object storage
	RestoreCompleted RestoreConditionType = "COMPLETED"
	RestoreFailed RestoreConditionType = "FAILED"
	// RestoreCanceled means the Restore operation was interrupted while being run
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


// RestoreCondition describes the observed state of a Restore at a certain point
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
	TimeCreated string `json:"timeCreated,omitempty"`
	TimeStarted string `json:"timeStarted,omitempty"`
	TimeCompleted string `json:"timeCompleted,omitempty"`
	Condition *RestoreCondition `json:"conditions,omitempty"`
	// Progress is a percentage, 100% means the operation is completed, either successfully or with errors
	Progress string `json:"progress,omitempty"`

	// Name of the pod the restore operation is executed on
	CoordinatorMember 		string `json:"coordinatorMember,omitempty"`

	Phase RestorationPhaseType `json:"restorationPhase,omitempty"`
	// unique identifier of an operation, a random id is assigned to each operation after a request is submitted,
	// from caller's perspective, an id is sent back as a response to his request so he can further query state of that operation,
	// referencing id, by operations/{id} endpoint
	ID string `json:"id,omitempty"`
}

// CassandraRestoreSpec defines the specification for a restore of a Cassandra backup.
type CassandraRestoreSpec struct {
	// Name of the CassandraCluster the restore belongs to
	CassandraCluster 		string `json:"cassandraCluster"`
	// Name of the CassandraBackup to restore
	CassandraBackup 		string `json:"cassandraBackup"`
	// Maximum number of threads used to download files from the cloud. Defaults to 10
	ConcurrentConnection 	*int32 `json:"concurrentConnection,omitempty"`
	// Directory of Cassandra where data folder resides. Defaults to /var/lib/cassandra
	CassandraDirectory 		string `json:"cassandraDirectory,omitempty"`
	// When set do not delete truncated SSTables after they've been restored during CLEANUP phase.
	// Defaults to false
	NoDeleteTruncates 		bool `json:"noDeleteTruncates,omitempty"`
	// Version of the schema to restore from. Upon backup, a schema version is automatically appended to a snapshot
	// name and its manifest is uploaded under that name. In case we have two snapshots having same name, we might
	// distinguish between the two of them by using the schema version. If schema version is not specified, we expect
	// a unique backup taken with respective snapshot name. This schema version has to match the version of a Cassandra
	// node we are doing restore for (hence, by proxy, when global request mode is used, all nodes have to be on exact
	// same schema version). Defaults to False
	SchemaVersion 			string `json:"schemaVersion,omitempty"`
	// When set a running node's schema version must match the snapshot's schema version. There might be cases when we
	// want to restore a table for which its CQL schema has not changed but it has changed for other table / keyspace
	// but a schema for that node has changed by doing that. Defaults to False
	ExactSchemaVersion 		bool `json:"exactSchemaVersion,omitempty""`
	// Strategy telling how we should go about restoration, please refer to details in backup and sidecar documentation
	// +kubebuilder:validation:Enum={"HARDLINKS","IMPORT"}
	RestorationStrategyType string `json:"restorationStrategyType,omitempty"`
	// Database entities to restore, it might be either only keyspaces or only tables prefixed by their respective
	// keyspace, e.g. 'k1,k2' if one wants to backup whole keyspaces or 'ks1.t1,ks2.t2' if one wants to restore specific
	// tables. These formats are mutually exclusive so 'k1,k2.t2' is invalid. An empty field will restore all keyspaces
	Entities 				string `json:"entities,omitempty"`
	// Name of Secret to use when accessing cloud storage providers
	Secret 					string `json:"secret,omitempty"`
}

// +genclient0
// +genclient:noStatus
// +resourceName=cassandrarestores
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraRestore is a Casskop Operator resource that represents the restoration of a backup of a Cassandra cluster
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

func ComputeRestorationStatus(restoreOperationReponse *csapi.RestoreOperationResponse) CassandraRestoreStatus{
	status := CassandraRestoreStatus{
		Progress:      ProgressPercentage(restoreOperationReponse.Progress),
		ID:            restoreOperationReponse.Id,
		TimeCreated:   restoreOperationReponse.CreationTime,
		TimeStarted:   restoreOperationReponse.StartTime,
		TimeCompleted: restoreOperationReponse.CompletionTime,
		Condition:     &RestoreCondition{},
		Phase:         RestorationPhase(restoreOperationReponse.RestorationPhase),
	}

	status.setRestorationCondition(restoreOperationReponse)

	return status
}

func RestorationPhase(phase string) RestorationPhaseType {
	restorationPhaseType := RestorationPhaseType(phase)
	switch restorationPhaseType {
	case RestorationPhaseDownload, RestorationPhaseImport, RestorationPhaseTruncate, RestorationPhaseCleanup:
		return restorationPhaseType
	default:
		return RestorePhaseUnknown
	}
}

func (status *CassandraRestoreStatus) setRestorationCondition(restore *csapi.RestoreOperationResponse) {
	status.Condition.LastTransitionTime = metav1.Now().Format(util.TimeStampLayout)
	status.Condition.Type = RestoreConditionType(restore.State)

	if status.Condition.Type.IsInError() {
		var failureCause FailureCause
		mapstructure.Decode(*restore.FailureCause, &failureCause)

		status.Condition.FailureCause = &failureCause
	}
}