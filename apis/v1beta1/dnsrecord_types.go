package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// DNSRecordSpec defines the desired state of DNSRecord
type DNSRecordSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider     DNSRecordParameters `json:"forProvider"`
}

// DNSRecordParameters are the configurable fields of a DNSRecord.
type DNSRecordParameters struct {
	// Domain is the domain name this DNS record belongs to
	// +kubebuilder:validation:Required
	Domain string `json:"domain"`

	// Type is the DNS record type (A, AAAA, CNAME, MX, TXT, SRV, etc.)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=A;AAAA;CNAME;MX;TXT;SRV;NS;PTR;CAA
	Type string `json:"type"`

	// Name is the record name (subdomain)
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Value is the record value
	// +kubebuilder:validation:Required
	Value string `json:"value"`

	// TTL is the time to live for the record in seconds
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:Maximum=86400
	// +optional
	TTL *int `json:"ttl,omitempty"`

	// Priority is used for MX and SRV records
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	Priority *int `json:"priority,omitempty"`

	// Weight is used for SRV records
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +optional
	Weight *int `json:"weight,omitempty"`

	// Port is used for SRV records
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	Port *int `json:"port,omitempty"`
}

// DNSRecordStatus defines the observed state of DNSRecord
type DNSRecordStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider        DNSRecordObservation `json:"atProvider,omitempty"`
}

// DNSRecordObservation are the observable fields of a DNSRecord.
type DNSRecordObservation struct {
	// ID is the unique identifier for the DNS record
	ID string `json:"id,omitempty"`

	// FQDN is the fully qualified domain name
	FQDN string `json:"fqdn,omitempty"`

	// CreatedDate is when the record was created
	CreatedDate *metav1.Time `json:"createdDate,omitempty"`

	// UpdatedDate is when the record was last updated
	UpdatedDate *metav1.Time `json:"updatedDate,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,namecheap}
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="VALUE",type="string",JSONPath=".spec.forProvider.value"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// DNSRecord is the Schema for the dnsrecords API
type DNSRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSRecordSpec   `json:"spec,omitempty"`
	Status DNSRecordStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DNSRecordList contains a list of DNSRecord
type DNSRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSRecord `json:"items"`
}

// GetCondition of this DNSRecord.
func (mg *DNSRecord) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this DNSRecord.
func (mg *DNSRecord) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this DNSRecord.
func (mg *DNSRecord) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this DNSRecord.
func (mg *DNSRecord) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this DNSRecord.
func (mg *DNSRecord) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this DNSRecord.
func (mg *DNSRecord) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this DNSRecord.
func (mg *DNSRecord) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this DNSRecord.
func (mg *DNSRecord) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this DNSRecord.
func (mg *DNSRecord) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this DNSRecord.
func (mg *DNSRecord) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this DNSRecord.
func (mg *DNSRecord) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this DNSRecord.
func (mg *DNSRecord) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

func init() {
	SchemeBuilder.Register(&DNSRecord{}, &DNSRecordList{})
}