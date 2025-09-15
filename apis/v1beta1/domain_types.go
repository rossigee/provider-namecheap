package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// DomainSpec defines the desired state of Domain
type DomainSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider     DomainParameters `json:"forProvider"`
}

// DomainParameters are the configurable fields of a Domain.
type DomainParameters struct {
	// DomainName is the domain name to manage
	// +kubebuilder:validation:Required
	DomainName string `json:"domainName"`

	// RegistrationYears specifies the number of years to register the domain for
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +optional
	RegistrationYears *int `json:"registrationYears,omitempty"`

	// RenewalYears specifies the number of years to renew the domain for
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +optional
	RenewalYears *int `json:"renewalYears,omitempty"`

	// Nameservers specifies custom nameservers for the domain
	// +optional
	Nameservers []string `json:"nameservers,omitempty"`

	// AutoRenew enables automatic domain renewal
	// +optional
	AutoRenew *bool `json:"autoRenew,omitempty"`

	// PrivacyProtection enables WHOIS privacy protection
	// +optional
	PrivacyProtection *bool `json:"privacyProtection,omitempty"`

	// WhoisGuardForwardEmail specifies the email address to forward WhoisGuard emails to
	// +optional
	WhoisGuardForwardEmail *string `json:"whoisGuardForwardEmail,omitempty"`
}

// DomainStatus defines the observed state of Domain
type DomainStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider        DomainObservation `json:"atProvider,omitempty"`
}

// DomainObservation are the observable fields of a Domain.
type DomainObservation struct {
	// ID is the unique identifier for the domain
	ID string `json:"id,omitempty"`

	// Status is the current status of the domain
	Status string `json:"status,omitempty"`

	// ExpirationDate is when the domain expires
	ExpirationDate *metav1.Time `json:"expirationDate,omitempty"`

	// CreatedDate is when the domain was created
	CreatedDate *metav1.Time `json:"createdDate,omitempty"`

	// UpdatedDate is when the domain was last updated
	UpdatedDate *metav1.Time `json:"updatedDate,omitempty"`

	// Nameservers are the current nameservers for the domain
	Nameservers []string `json:"nameservers,omitempty"`

	// IsExpired indicates if the domain has expired
	IsExpired *bool `json:"isExpired,omitempty"`

	// IsLocked indicates if the domain is locked
	IsLocked *bool `json:"isLocked,omitempty"`

	// IsAutoRenew indicates if auto-renewal is enabled
	IsAutoRenew *bool `json:"isAutoRenew,omitempty"`

	// WhoisGuardStatus indicates the current WhoisGuard status
	WhoisGuardStatus *string `json:"whoisGuardStatus,omitempty"`

	// WhoisGuardID is the WhoisGuard service ID
	WhoisGuardID *int `json:"whoisGuardID,omitempty"`

	// IsPremium indicates if this is a premium domain
	IsPremium *bool `json:"isPremium,omitempty"`

	// IsOurDNS indicates if using Namecheap DNS hosting
	IsOurDNS *bool `json:"isOurDNS,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,namecheap}
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// Domain is the Schema for the domains API
type Domain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DomainSpec   `json:"spec,omitempty"`
	Status DomainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DomainList contains a list of Domain
type DomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Domain `json:"items"`
}

// GetCondition of this Domain.
func (mg *Domain) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this Domain.
func (mg *Domain) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this Domain.
func (mg *Domain) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this Domain.
func (mg *Domain) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this Domain.
func (mg *Domain) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this Domain.
func (mg *Domain) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this Domain.
func (mg *Domain) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this Domain.
func (mg *Domain) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this Domain.
func (mg *Domain) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this Domain.
func (mg *Domain) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this Domain.
func (mg *Domain) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this Domain.
func (mg *Domain) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

func init() {
	SchemeBuilder.Register(&Domain{}, &DomainList{})
}