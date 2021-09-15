package apis

import (
	"github.com/Orange-OpenSource/casskop/multi-casskop/pkg/apis/db/v2"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v2.SchemeBuilder.AddToScheme)
}
