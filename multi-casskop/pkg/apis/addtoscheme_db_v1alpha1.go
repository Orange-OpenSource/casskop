package apis

import (
	"github.com/Orange-OpenSource/cassandra-k8s-operator/multi-casskop/pkg/apis/db/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}
