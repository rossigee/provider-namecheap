// Package v1beta1 contains API Schema definitions for the namecheap v1beta1 API group
// +kubebuilder:object:generate=true
// +groupName=namecheap.m.crossplane.io
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group   = "namecheap.m.crossplane.io"
	Version = "v1beta1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion,
		&Domain{},
		&DomainList{},
		&ProviderConfig{},
		&ProviderConfigList{},
		&ProviderConfigUsage{},
		&ProviderConfigUsageList{},
	)
	return nil
}
