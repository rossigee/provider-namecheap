package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A ProviderConfigUsage indicates that a resource is using a ProviderConfig.
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONFIG-NAME",type="string",JSONPath=".providerConfigRef.name"
// +kubebuilder:printcolumn:name="RESOURCE-KIND",type="string",JSONPath=".resourceRef.kind"
// +kubebuilder:printcolumn:name="RESOURCE-NAME",type="string",JSONPath=".resourceRef.name"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,namecheap}
// +kubebuilder:object:root=true
type ProviderConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	xpv1.TypedProviderConfigUsage `json:",inline"`
}

// GetProviderConfigReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) GetProviderConfigReference() xpv1.ProviderConfigReference {
	return mg.ProviderConfigReference
}

// SetProviderConfigReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) SetProviderConfigReference(r xpv1.ProviderConfigReference) {
	mg.ProviderConfigReference = r
}

// GetResourceReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) GetResourceReference() xpv1.TypedReference {
	return mg.ResourceReference
}

// SetResourceReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) SetResourceReference(r xpv1.TypedReference) {
	mg.ResourceReference = r
}

// ProviderConfigUsageList contains a list of ProviderConfigUsage
// +kubebuilder:object:root=true
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}
