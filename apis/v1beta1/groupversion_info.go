// Package v1beta1 contains API Schema definitions for the namecheap v1beta1 API group
// +kubebuilder:object:generate=true
// +groupName=namecheap.m.crossplane.io
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "namecheap.m.crossplane.io", Version: "v1beta1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(s *runtime.Scheme) error {
	return nil
}
