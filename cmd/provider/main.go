package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"

	"github.com/rossigee/provider-namecheap/apis"
	"github.com/rossigee/provider-namecheap/internal/controller/domain"
	"github.com/rossigee/provider-namecheap/internal/controller/dnsrecord"
)

func main() {
	var (
		app                     = kingpin.New(filepath.Base(os.Args[0]), "Crossplane provider for Namecheap").DefaultEnvars()
		debug                   = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		syncInterval            = app.Flag("sync", "Sync interval controls how often all resources will be double checked for drift.").Short('s').Default("1h").Duration()
		pollInterval            = app.Flag("poll", "Poll interval controls how often an individual resource should be checked for drift.").Default("1m").Duration()
		leaderElection          = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").Bool()
		maxReconcileRate        = app.Flag("max-reconcile-rate", "The global maximum rate per second at which resources may checked for drift from the desired state.").Default("100").Int()
		namespace               = app.Flag("namespace", "Namespace used to set as default scope in default secret store config.").Default("crossplane-system").String()
		enableExternalSecretStores = app.Flag("enable-external-secret-stores", "Enable support for external secret stores.").Default("false").Bool()
		enableManagementPolicies   = app.Flag("enable-management-policies", "Enable support for Management Policies.").Default("true").Bool()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-namecheap"))
	if *debug {
		// The controller-runtime runs with a no-op logger by default. It is
		// *very* verbose even at info level, so we only provide it a real
		// logger when we're running in debug mode.
		ctrl.SetLogger(zl)
	}

	// currently, we configure the jitter to be the 5% of the poll interval
	pollJitter := time.Duration(float64(*pollInterval) * 0.05)
	log.Debug("Starting", "sync-interval", syncInterval.String(),
		"poll-interval", pollInterval.String(), "poll-jitter", pollJitter, "max-reconcile-rate", *maxReconcileRate)

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	// Get the namespace for the leader election
	leaderElectionNamespace := ""
	if *leaderElection {
		leaderElectionNamespace = *namespace
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		LeaderElection:          *leaderElection,
		LeaderElectionID:        "crossplane-leader-election-provider-namecheap",
		LeaderElectionNamespace: leaderElectionNamespace,
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
		Cache: cache.Options{
			SyncPeriod: syncInterval,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			CertDir: os.Getenv("WEBHOOK_TLS_CERT_DIR"),
		}),
		Metrics: server.Options{
			BindAddress: ":8080",
		},
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	featureFlags := &feature.Flags{}
	o := controller.Options{
		Logger:                  log,
		MaxConcurrentReconciles: *maxReconcileRate,
		PollInterval:            *pollInterval,
		GlobalRateLimiter:       ratelimiter.NewGlobal(*maxReconcileRate),
		Features:                featureFlags,
	}

	if *enableExternalSecretStores {
		// External secret stores feature would be enabled here
		log.Info("External secret stores feature requested but not implemented")
	}

	if *enableManagementPolicies {
		featureFlags.Enable(feature.EnableBetaManagementPolicies)
		log.Info("Beta feature enabled", "flag", feature.EnableBetaManagementPolicies)
	}

	kingpin.FatalIfError(apis.AddToScheme(mgr.GetScheme()), "Cannot add Namecheap APIs to scheme")

	kingpin.FatalIfError(domain.Setup(mgr, o), "Cannot setup Domain controller")
	kingpin.FatalIfError(dnsrecord.Setup(mgr, o), "Cannot setup DNSRecord controller")

	ctx := ctrl.SetupSignalHandler()
	kingpin.FatalIfError(mgr.Start(ctx), "Cannot start controller manager")
}