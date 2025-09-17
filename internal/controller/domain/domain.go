package domain

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
	errNotDomain    = "managed resource is not a Domain custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient        = "cannot create new Service"
	errCreateDomain     = "cannot create domain"
	errUpdateDomain     = "cannot update domain"
	errDeleteDomain     = "cannot delete domain"
	errGetDomain        = "cannot get domain"
	errSetNameservers   = "cannot set nameservers"
)

// Setup adds a controller that reconciles Domain managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.DomainGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.DomainGroupVersionKind),
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
		For(&v1beta1.Domain{}).
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
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return nil, errors.New(errNotDomain)
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

	if pc.Spec.APIBase != nil {
		config.BaseURL = *pc.Spec.APIBase
	}

	client := namecheap.NewClient(config)

	return &external{client: client}, nil
}

// Disconnect cleans up any resources created by Connect.
func (c *external) Disconnect(ctx context.Context) error {
	// No cleanup needed for HTTP client
	return nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client *namecheap.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDomain)
	}

	domainName := cr.Spec.ForProvider.DomainName
	if domainName == "" {
		return managed.ExternalObservation{}, nil
	}

	// Check if domain exists
	exists, err := c.client.DomainExists(ctx, domainName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetDomain)
	}

	if !exists {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Get domain details
	domain, err := c.client.GetDomain(ctx, domainName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetDomain)
	}

	// Update status with observed values
	cr.Status.AtProvider.ID = strconv.Itoa(domain.ID)
	cr.Status.AtProvider.Status = "Active" // Namecheap doesn't provide status in API response
	if !domain.Created.IsZero() {
		cr.Status.AtProvider.CreatedDate = &metav1.Time{Time: domain.Created}
	}
	if !domain.Expires.IsZero() {
		cr.Status.AtProvider.ExpirationDate = &metav1.Time{Time: domain.Expires}
	}

	// Set external name annotation
	meta.SetExternalName(cr, domainName)

	// Check if resource is up to date
	upToDate := true

	// Check nameservers if specified
	// Note: Nameserver comparison would require additional API call
	// For now, we assume nameservers are up to date if domain exists

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDomain)
	}

	cr.Status.SetConditions(xpv1.Creating())

	domainName := cr.Spec.ForProvider.DomainName
	years := 1
	if cr.Spec.ForProvider.RegistrationYears != nil {
		years = *cr.Spec.ForProvider.RegistrationYears
	}

	// Create the domain
	domain, err := c.client.CreateDomain(ctx, domainName, years)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateDomain)
	}

	// Set external name
	meta.SetExternalName(cr, domainName)

	// Update status
	cr.Status.AtProvider.ID = strconv.Itoa(domain.ID)

	// Set nameservers if specified
	if len(cr.Spec.ForProvider.Nameservers) > 0 {
		if err := c.client.SetNameservers(ctx, domainName, cr.Spec.ForProvider.Nameservers); err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errSetNameservers)
		}
	}

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDomain)
	}

	domainName := cr.Spec.ForProvider.DomainName

	// Handle domain renewal if requested
	if cr.Spec.ForProvider.RenewalYears != nil {
		years := *cr.Spec.ForProvider.RenewalYears
		_, err := c.client.RenewDomain(ctx, domainName, years)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot renew domain")
		}
		// Clear the renewal request after successful renewal
		cr.Spec.ForProvider.RenewalYears = nil
	}

	// Handle WhoisGuard privacy protection
	if cr.Spec.ForProvider.PrivacyProtection != nil {
		whoisGuard, err := c.client.GetWhoisGuardForDomain(ctx, domainName)
		enabled := *cr.Spec.ForProvider.PrivacyProtection

		if err == nil {
			// WhoisGuard exists, check if we need to enable/disable it
			currentlyEnabled := whoisGuard.Status == "ENABLED"

			if enabled && !currentlyEnabled {
				// Enable WhoisGuard
				forwardEmail := ""
				if cr.Spec.ForProvider.WhoisGuardForwardEmail != nil {
					forwardEmail = *cr.Spec.ForProvider.WhoisGuardForwardEmail
				}
				if err := c.client.EnableWhoisGuard(ctx, whoisGuard.ID, domainName, forwardEmail); err != nil {
					return managed.ExternalUpdate{}, errors.Wrap(err, "cannot enable WhoisGuard")
				}
			} else if !enabled && currentlyEnabled {
				// Disable WhoisGuard
				if err := c.client.DisableWhoisGuard(ctx, whoisGuard.ID, domainName); err != nil {
					return managed.ExternalUpdate{}, errors.Wrap(err, "cannot disable WhoisGuard")
				}
			}
		}
	}

	// Update nameservers if specified
	if len(cr.Spec.ForProvider.Nameservers) > 0 {
		if err := c.client.SetNameservers(ctx, domainName, cr.Spec.ForProvider.Nameservers); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errSetNameservers)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotDomain)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	// Note: Namecheap doesn't support domain deletion via API
	// Domains remain in the account but cannot be programmatically deleted
	// This is a limitation of the Namecheap API

	return managed.ExternalDelete{}, nil
}