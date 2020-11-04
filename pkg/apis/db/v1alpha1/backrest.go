package v1alpha1

import (
	"fmt"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	"github.com/mitchellh/mapstructure"
	"strconv"
)

func ProgressPercentage(progress float64) string {
	return fmt.Sprintf("%v%%", strconv.Itoa(int(progress*100)))
}

func failureCause(errors []icarus.ErrorObject) []FailureCause {
	var failureCause []FailureCause
	mapstructure.Decode(errors, &failureCause)
	return failureCause
}

// BackRestCondition describes the observed state of a Restore at a certain point
type BackRestCondition struct {
	Type string `json:"type"`
	// +optional
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// +optional
	FailureCause []FailureCause `json:"failureCause,omitempty"`
}

type BackRestStatus struct {
	TimeCreated   string             `json:"timeCreated,omitempty"`
	TimeStarted   string             `json:"timeStarted,omitempty"`
	TimeCompleted string             `json:"timeCompleted,omitempty"`
	Condition     *BackRestCondition `json:"condition,omitempty"`
	// Name of the pod the restore operation is executed on
	CoordinatorMember string `json:"coordinatorMember,omitempty"`
	// Progress is a percentage, 100% means the operation is completed, either successfully or with errors
	Progress string `json:"progress,omitempty"`
	// unique identifier of an operation, a random id is assigned to each operation after a request is submitted,
	// from caller's perspective, an id is sent back as a response to his request so he can further query state of that operation,
	// referencing id, by operations/{id} endpoint
	ID string `json:"id,omitempty"`
}

type FailureCause struct {
	// hostame of a node where this error has occurred
	Source string `json:"source,omitempty"`
	// message explaining the error
	Message string `json:"message,omitempty"`
}

