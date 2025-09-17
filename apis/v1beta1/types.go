package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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

	xpv1.ProviderConfigUsage `json:",inline"`
}

// GetProviderConfigReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) GetProviderConfigReference() xpv1.Reference {
	return mg.ProviderConfigUsage.ProviderConfigReference
}

// SetProviderConfigReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) SetProviderConfigReference(r xpv1.Reference) {
	mg.ProviderConfigUsage.ProviderConfigReference = r
}

// GetResourceReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) GetResourceReference() xpv1.TypedReference {
	return mg.ProviderConfigUsage.ResourceReference
}

// SetResourceReference of this ProviderConfigUsage.
func (mg *ProviderConfigUsage) SetResourceReference(r xpv1.TypedReference) {
	mg.ProviderConfigUsage.ResourceReference = r
}

// ProviderConfigUsageList contains a list of ProviderConfigUsage
// +kubebuilder:object:root=true
type ProviderConfigUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfigUsage `json:"items"`
}


func init() {
	SchemeBuilder.Register(&ProviderConfigUsage{}, &ProviderConfigUsageList{})
}