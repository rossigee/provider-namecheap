package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SSLCertificateSpec defines the desired state of SSLCertificate
type SSLCertificateSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SSLCertificateParameters `json:"forProvider"`
}

// SSLCertificateParameters are the configurable fields of an SSLCertificate.
type SSLCertificateParameters struct {
	// CertificateType specifies the type of SSL certificate to purchase
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	CertificateType int `json:"certificateType"`

	// Years specifies the number of years to purchase the certificate for
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3
	// +kubebuilder:default=1
	// +optional
	Years *int `json:"years,omitempty"`

	// SANsToAdd specifies additional Subject Alternative Names
	// +optional
	SANsToAdd *string `json:"sansToAdd,omitempty"`

	// DomainName is the primary domain name for the certificate
	// +kubebuilder:validation:Required
	DomainName string `json:"domainName"`

	// CSR is the Certificate Signing Request
	// +optional
	CSR *string `json:"csr,omitempty"`

	// ApproverEmail is the email address for certificate approval
	// +optional
	ApproverEmail *string `json:"approverEmail,omitempty"`

	// HTTPDCValidation enables HTTP domain control validation
	// +optional
	HTTPDCValidation *string `json:"httpDCValidation,omitempty"`

	// DNSValidation enables DNS domain control validation
	// +optional
	DNSValidation *string `json:"dnsValidation,omitempty"`

	// WebServerType specifies the web server type for certificate format
	// +optional
	WebServerType *string `json:"webServerType,omitempty"`

	// AutoActivate automatically activates the certificate after purchase
	// +optional
	AutoActivate *bool `json:"autoActivate,omitempty"`
}

// SSLCertificateStatus defines the observed state of SSLCertificate
type SSLCertificateStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SSLCertificateObservation `json:"atProvider,omitempty"`
}

// SSLCertificateObservation are the observable fields of an SSLCertificate.
type SSLCertificateObservation struct {
	// CertificateID is the unique identifier for the SSL certificate
	CertificateID *int `json:"certificateID,omitempty"`

	// HostName is the hostname the certificate is issued for
	HostName *string `json:"hostName,omitempty"`

	// SSLType is the type of SSL certificate
	SSLType *string `json:"sslType,omitempty"`

	// PurchaseDate is when the certificate was purchased
	PurchaseDate *metav1.Time `json:"purchaseDate,omitempty"`

	// ExpireDate is when the certificate expires
	ExpireDate *metav1.Time `json:"expireDate,omitempty"`

	// ActivationExpireDate is when the activation expires
	ActivationExpireDate *metav1.Time `json:"activationExpireDate,omitempty"`

	// IsExpired indicates if the certificate has expired
	IsExpired *bool `json:"isExpired,omitempty"`

	// Status is the current status of the certificate
	Status *string `json:"status,omitempty"`

	// StatusDescription provides detailed status information
	StatusDescription *string `json:"statusDescription,omitempty"`

	// Years is the number of years the certificate was purchased for
	Years *int `json:"years,omitempty"`

	// OrderID is the order identifier
	OrderID *int `json:"orderID,omitempty"`

	// TransactionID is the transaction identifier
	TransactionID *int `json:"transactionID,omitempty"`

	// ChargedAmount is the amount charged for the certificate
	ChargedAmount *string `json:"chargedAmount,omitempty"`

	// Provider information
	ProviderName *string `json:"providerName,omitempty"`

	// ApproverEmailList contains valid approver email addresses
	ApproverEmailList []string `json:"approverEmailList,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,namecheap}
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="HOSTNAME",type="string",JSONPath=".status.atProvider.hostName"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// SSLCertificate is the Schema for the sslcertificates API
type SSLCertificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSLCertificateSpec   `json:"spec,omitempty"`
	Status SSLCertificateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SSLCertificateList contains a list of SSLCertificate
type SSLCertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSLCertificate `json:"items"`
}

// GetCondition of this SSLCertificate.
func (mg *SSLCertificate) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this SSLCertificate.
func (mg *SSLCertificate) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this SSLCertificate.
func (mg *SSLCertificate) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this SSLCertificate.
func (mg *SSLCertificate) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this SSLCertificate.
func (mg *SSLCertificate) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this SSLCertificate.
func (mg *SSLCertificate) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this SSLCertificate.
func (mg *SSLCertificate) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this SSLCertificate.
func (mg *SSLCertificate) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this SSLCertificate.
func (mg *SSLCertificate) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this SSLCertificate.
func (mg *SSLCertificate) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this SSLCertificate.
func (mg *SSLCertificate) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this SSLCertificate.
func (mg *SSLCertificate) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

func init() {
	SchemeBuilder.Register(&SSLCertificate{}, &SSLCertificateList{})
}