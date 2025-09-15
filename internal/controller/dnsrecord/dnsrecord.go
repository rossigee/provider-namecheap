package dnsrecord

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
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
	"github.com/crossplane/crossplane-runtime/pkg/resource/fake"

	"github.com/rossigee/provider-namecheap/apis/v1beta1"
	"github.com/rossigee/provider-namecheap/internal/clients/namecheap"
)

const (
	errNotDNSRecord = "managed resource is not a DNSRecord custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient         = "cannot create new Service"
	errCreateDNSRecord   = "cannot create DNS record"
	errUpdateDNSRecord   = "cannot update DNS record"
	errDeleteDNSRecord   = "cannot delete DNS record"
	errGetDNSRecord      = "cannot get DNS record"
)

// Setup adds a controller that reconciles DNSRecord managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.DNSRecordGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.DNSRecordGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &fake.ProviderConfigUsage{}),
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.DNSRecord{}).
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
	cr, ok := mg.(*v1beta1.DNSRecord)
	if !ok {
		return nil, errors.New(errNotDNSRecord)
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
		APIUser  string `json:"apiUser"`
		APIKey   string `json:"apiKey"`
		Username string `json:"username"`
		ClientIP string `json:"clientIP"`
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
	cr, ok := mg.(*v1beta1.DNSRecord)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDNSRecord)
	}

	domain := cr.Spec.ForProvider.Domain
	recordName := cr.Spec.ForProvider.Name
	recordType := cr.Spec.ForProvider.Type

	if domain == "" || recordName == "" || recordType == "" {
		return managed.ExternalObservation{}, nil
	}

	// Check if DNS record exists
	exists, err := c.client.DNSRecordExists(ctx, domain, recordName, recordType)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetDNSRecord)
	}

	if !exists {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Get DNS record details
	record, err := c.client.GetDNSRecord(ctx, domain, recordName, recordType)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetDNSRecord)
	}

	// Update status with observed values
	cr.Status.AtProvider.ID = strconv.Itoa(record.HostID)
	cr.Status.AtProvider.FQDN = recordName + "." + domain

	// Set external name annotation
	externalName := domain + "/" + recordType + "/" + recordName
	meta.SetExternalName(cr, externalName)

	// Check if resource is up to date
	upToDate := true
	if record.Address != cr.Spec.ForProvider.Value {
		upToDate = false
	}
	if cr.Spec.ForProvider.TTL != nil && record.TTL != *cr.Spec.ForProvider.TTL {
		upToDate = false
	}
	if cr.Spec.ForProvider.Priority != nil && record.MXPref != *cr.Spec.ForProvider.Priority {
		upToDate = false
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.DNSRecord)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDNSRecord)
	}

	cr.Status.SetConditions(xpv1.Creating())

	domain := cr.Spec.ForProvider.Domain
	recordName := cr.Spec.ForProvider.Name
	recordType := cr.Spec.ForProvider.Type
	recordValue := cr.Spec.ForProvider.Value

	// Create DNS record struct
	record := namecheap.DNSRecord{
		Name:    recordName,
		Type:    recordType,
		Address: recordValue,
		TTL:     300, // Default TTL
	}

	if cr.Spec.ForProvider.TTL != nil {
		record.TTL = *cr.Spec.ForProvider.TTL
	}

	if cr.Spec.ForProvider.Priority != nil {
		record.MXPref = *cr.Spec.ForProvider.Priority
	}

	// Create the DNS record
	if err := c.client.CreateDNSRecord(ctx, domain, record); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateDNSRecord)
	}

	// Set external name
	externalName := domain + "/" + recordType + "/" + recordName
	meta.SetExternalName(cr, externalName)

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.DNSRecord)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDNSRecord)
	}

	domain := cr.Spec.ForProvider.Domain
	recordName := cr.Spec.ForProvider.Name
	recordType := cr.Spec.ForProvider.Type
	recordValue := cr.Spec.ForProvider.Value

	// Get existing record to preserve HostID
	existingRecord, err := c.client.GetDNSRecord(ctx, domain, recordName, recordType)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetDNSRecord)
	}

	// Update DNS record struct
	record := namecheap.DNSRecord{
		HostID:  existingRecord.HostID,
		Name:    recordName,
		Type:    recordType,
		Address: recordValue,
		TTL:     300, // Default TTL
	}

	if cr.Spec.ForProvider.TTL != nil {
		record.TTL = *cr.Spec.ForProvider.TTL
	}

	if cr.Spec.ForProvider.Priority != nil {
		record.MXPref = *cr.Spec.ForProvider.Priority
	}

	// Update the DNS record
	if err := c.client.UpdateDNSRecord(ctx, domain, record); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateDNSRecord)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.DNSRecord)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotDNSRecord)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	domain := cr.Spec.ForProvider.Domain
	recordName := cr.Spec.ForProvider.Name
	recordType := cr.Spec.ForProvider.Type

	// Delete the DNS record
	if err := c.client.DeleteDNSRecord(ctx, domain, recordName, recordType); err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteDNSRecord)
	}

	return managed.ExternalDelete{}, nil
}