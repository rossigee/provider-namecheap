package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/rossigee/provider-namecheap/apis/v1beta1"
)

func init() {
	// AddToSchemes may be used to add all resources defined in the project to a Scheme
	AddToSchemes = append(AddToSchemes, v1beta1.SchemeBuilder.AddToScheme)
}

// AddToSchemes is a global list of functions to add items to a scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}