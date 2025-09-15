package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group   = "namecheap.m.crossplane.io"
	Version = "v1beta1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// Domain
	DomainKind             = "Domain"
	DomainGroupKind        = schema.GroupKind{Group: Group, Kind: DomainKind}.String()
	DomainKindAPIVersion   = DomainKind + "." + SchemeGroupVersion.String()
	DomainGroupVersionKind = SchemeGroupVersion.WithKind(DomainKind)

	// DNSRecord
	DNSRecordKind             = "DNSRecord"
	DNSRecordGroupKind        = schema.GroupKind{Group: Group, Kind: DNSRecordKind}.String()
	DNSRecordKindAPIVersion   = DNSRecordKind + "." + SchemeGroupVersion.String()
	DNSRecordGroupVersionKind = SchemeGroupVersion.WithKind(DNSRecordKind)

	// ProviderConfig
	ProviderConfigKind             = "ProviderConfig"
	ProviderConfigGroupKind        = schema.GroupKind{Group: Group, Kind: ProviderConfigKind}.String()
	ProviderConfigKindAPIVersion   = ProviderConfigKind + "." + SchemeGroupVersion.String()
	ProviderConfigGroupVersionKind = SchemeGroupVersion.WithKind(ProviderConfigKind)

	// ProviderConfigUsage
	ProviderConfigUsageKind             = "ProviderConfigUsage"
	ProviderConfigUsageGroupKind        = schema.GroupKind{Group: Group, Kind: ProviderConfigUsageKind}.String()
	ProviderConfigUsageKindAPIVersion   = ProviderConfigUsageKind + "." + SchemeGroupVersion.String()
	ProviderConfigUsageGroupVersionKind = SchemeGroupVersion.WithKind(ProviderConfigUsageKind)

	// SSLCertificate
	SSLCertificateKind             = "SSLCertificate"
	SSLCertificateGroupKind        = schema.GroupKind{Group: Group, Kind: SSLCertificateKind}.String()
	SSLCertificateKindAPIVersion   = SSLCertificateKind + "." + SchemeGroupVersion.String()
	SSLCertificateGroupVersionKind = SchemeGroupVersion.WithKind(SSLCertificateKind)
)

// ProviderConfigUsage tracks the usage of a ProviderConfig.
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,namecheap}
type ProviderConfigUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	ProviderConfigReference ProviderConfigReference `json:"providerConfigRef"`
	ResourceReference       ResourceReference       `json:"resourceRef"`
}

// GetProviderConfigReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) GetProviderConfigReference() ProviderConfigReference {
	return mg.ProviderConfigReference
}

// GetResourceReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) GetResourceReference() ResourceReference {
	return mg.ResourceReference
}

// SetProviderConfigReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) SetProviderConfigReference(r ProviderConfigReference) {
	mg.ProviderConfigReference = r
}

// SetResourceReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) SetResourceReference(r ResourceReference) {
	mg.ResourceReference = r
}

// ProviderConfigUsageList contains a list of ProviderConfigUsage
// +kubebuilder:object:root=true
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}

// ProviderConfigReference to the provider config being used.
type ProviderConfigReference struct {
	// Name of the referenced object.
	Name string `json:"name"`
}

// ResourceReference to the managed resource using the provider config.
type ResourceReference struct {
	// Name of the referenced object.
	Name string `json:"name"`

	// Namespace of the referenced object.
	Namespace string `json:"namespace,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ProviderConfigUsage{}, &ProviderConfigUsageList{})
}