package sslcertificate

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-namecheap/apis/v1beta1"
	"github.com/rossigee/provider-namecheap/internal/clients/namecheap"
)

const (
	errNotSSLCertificate   = "managed resource is not an SSLCertificate custom resource"
	errTrackPCUsage       = "cannot track ProviderConfig usage"
	errGetPC              = "cannot get ProviderConfig"
	errGetCreds           = "cannot get credentials"
	errNewClient          = "cannot create new Service"
	errGetSSLCertificate  = "cannot get SSL certificate"
	errCreateSSLCertificate = "cannot create SSL certificate"
	errActivateSSLCertificate = "cannot activate SSL certificate"
	errDeleteSSLCertificate = "cannot delete SSL certificate"
)

// Setup adds a controller that reconciles SSLCertificate managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.SSLCertificateGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.SSLCertificateGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.SSLCertificate{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube  client.Client
	usage resource.Tracker
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.SSLCertificate)
	if !ok {
		return nil, errors.New(errNotSSLCertificate)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &v1beta1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	// Parse credentials from the secret data
	var creds struct {
		APIUser  string `json:"api_user"`
		APIKey   string `json:"api_key"`
		Username string `json:"username"`
		ClientIP string `json:"client_ip"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, errors.Wrap(err, "failed to parse credentials JSON")
	}

	// Create Namecheap client
	config := namecheap.Config{
		APIUser:  creds.APIUser,
		APIKey:   creds.APIKey,
		Username: creds.Username,
		ClientIP: creds.ClientIP,
		Sandbox:  pc.Spec.SandboxMode != nil && *pc.Spec.SandboxMode,
	}

	client := namecheap.NewClient(config)

	return &external{service: client}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *namecheap.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.SSLCertificate)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSSLCertificate)
	}

	// If we don't have a certificate ID, the resource doesn't exist yet
	if cr.Status.AtProvider.CertificateID == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	certificateID := *cr.Status.AtProvider.CertificateID
	cert, err := c.service.GetSSLCertificate(ctx, certificateID)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetSSLCertificate)
	}

	// Update the status with observed values
	cr.Status.AtProvider.CertificateID = &cert.CommandResponse.SSLGetInfoResult.CertificateID
	cr.Status.AtProvider.HostName = &cert.CommandResponse.SSLGetInfoResult.HostName
	cr.Status.AtProvider.SSLType = &cert.CommandResponse.SSLGetInfoResult.SSLType
	cr.Status.AtProvider.IsExpired = &cert.CommandResponse.SSLGetInfoResult.IsExpiredYN
	cr.Status.AtProvider.Status = &cert.CommandResponse.SSLGetInfoResult.Status
	cr.Status.AtProvider.StatusDescription = &cert.CommandResponse.SSLGetInfoResult.StatusDescription
	cr.Status.AtProvider.Years = &cert.CommandResponse.SSLGetInfoResult.Years

	if !cert.CommandResponse.SSLGetInfoResult.PurchaseDate.IsZero() {
		cr.Status.AtProvider.PurchaseDate = &metav1.Time{Time: cert.CommandResponse.SSLGetInfoResult.PurchaseDate}
	}
	if !cert.CommandResponse.SSLGetInfoResult.ExpireDate.IsZero() {
		cr.Status.AtProvider.ExpireDate = &metav1.Time{Time: cert.CommandResponse.SSLGetInfoResult.ExpireDate}
	}
	if !cert.CommandResponse.SSLGetInfoResult.ActivationExpireDate.IsZero() {
		cr.Status.AtProvider.ActivationExpireDate = &metav1.Time{Time: cert.CommandResponse.SSLGetInfoResult.ActivationExpireDate}
	}

	cr.Status.AtProvider.ProviderName = &cert.CommandResponse.SSLGetInfoResult.Provider.Name
	cr.Status.AtProvider.ApproverEmailList = cert.CommandResponse.SSLGetInfoResult.ApproverEmailList

	// Set resource as ready if certificate is active
	if cert.CommandResponse.SSLGetInfoResult.Status == "ACTIVE" {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.SSLCertificate)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSSLCertificate)
	}

	years := 1
	if cr.Spec.ForProvider.Years != nil {
		years = *cr.Spec.ForProvider.Years
	}

	sansToAdd := ""
	if cr.Spec.ForProvider.SANsToAdd != nil {
		sansToAdd = *cr.Spec.ForProvider.SANsToAdd
	}

	certificateID, err := c.service.CreateSSLCertificate(ctx, cr.Spec.ForProvider.CertificateType, years, sansToAdd)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateSSLCertificate)
	}

	// Store the certificate ID
	cr.Status.AtProvider.CertificateID = &certificateID

	// Set external name annotation
	meta.SetExternalName(cr, strconv.Itoa(certificateID))

	// Auto-activate if requested and CSR is provided
	if cr.Spec.ForProvider.AutoActivate != nil && *cr.Spec.ForProvider.AutoActivate &&
		cr.Spec.ForProvider.CSR != nil && cr.Spec.ForProvider.ApproverEmail != nil {

		httpDCValidation := ""
		if cr.Spec.ForProvider.HTTPDCValidation != nil {
			httpDCValidation = *cr.Spec.ForProvider.HTTPDCValidation
		}

		dnsValidation := ""
		if cr.Spec.ForProvider.DNSValidation != nil {
			dnsValidation = *cr.Spec.ForProvider.DNSValidation
		}

		webServerType := ""
		if cr.Spec.ForProvider.WebServerType != nil {
			webServerType = *cr.Spec.ForProvider.WebServerType
		}

		err = c.service.ActivateSSLCertificate(ctx, certificateID, *cr.Spec.ForProvider.CSR,
			cr.Spec.ForProvider.DomainName, *cr.Spec.ForProvider.ApproverEmail,
			httpDCValidation, dnsValidation, webServerType)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errActivateSSLCertificate)
		}
	}

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{
			"certificate_id": []byte(strconv.Itoa(certificateID)),
			"domain_name":    []byte(cr.Spec.ForProvider.DomainName),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.SSLCertificate)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSSLCertificate)
	}

	// SSL certificates are mostly read-only after creation
	// The main updates would be reissuing or resending approval emails
	// These would be triggered by annotations or specific fields

	certificateID := *cr.Status.AtProvider.CertificateID

	// Check for reissue annotation
	if cr.Annotations != nil {
		if _, exists := cr.Annotations["namecheap.crossplane.io/reissue"]; exists {
			if cr.Spec.ForProvider.CSR != nil && cr.Spec.ForProvider.ApproverEmail != nil {
				err := c.service.ReissueSSLCertificate(ctx, certificateID, *cr.Spec.ForProvider.CSR, *cr.Spec.ForProvider.ApproverEmail)
				if err != nil {
					return managed.ExternalUpdate{}, errors.Wrap(err, "cannot reissue SSL certificate")
				}
				// Remove the annotation after successful reissue
				delete(cr.Annotations, "namecheap.crossplane.io/reissue")
			}
		}

		// Check for resend approval email annotation
		if _, exists := cr.Annotations["namecheap.crossplane.io/resend-approval"]; exists {
			err := c.service.ResendSSLApprovalEmail(ctx, certificateID)
			if err != nil {
				return managed.ExternalUpdate{}, errors.Wrap(err, "cannot resend SSL approval email")
			}
			// Remove the annotation after successful resend
			delete(cr.Annotations, "namecheap.crossplane.io/resend-approval")
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.SSLCertificate)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotSSLCertificate)
	}

	// SSL certificates cannot be deleted via API - they simply expire
	// We'll just mark the resource as being deleted
	cr.SetConditions(xpv1.Deleting())

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	// No persistent connection to close
	return nil
}