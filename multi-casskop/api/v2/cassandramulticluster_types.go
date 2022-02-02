package v2

import (
	apicc "github.com/Orange-OpenSource/casskop/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MultiCasskopSpec defines the desired state of MultiCasskop
// +k8s:openapi-gen=true
type MultiCasskopSpec struct {
	DeleteCassandraCluster *bool                             `json:"deleteCassandraCluster,omitempty"`
	Base                   apicc.CassandraCluster            `json:"base,omitempty"`
	Override               map[string]apicc.CassandraCluster `json:"override,omitempty"`
}

// MultiCasskopStatus defines the observed state of MultiCasskop
// +k8s:openapi-gen=true
type MultiCasskopStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +kubebuilder:object:root=true

// MultiCasskop is the Schema for the MultiCasskops API
// +k8s:openapi-gen=true
type MultiCasskop struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultiCasskopSpec   `json:"spec,omitempty"`
	Status MultiCasskopStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MultiCasskopList contains a list of MultiCasskop
type MultiCasskopList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultiCasskop `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MultiCasskop{}, &MultiCasskopList{})
}
