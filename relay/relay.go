package relay

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"

	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/pokt-foundation/pocket-go/relayer"
	"github.com/pokt-foundation/pocket-go/signer"

	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/pokt-foundation/portal-api-go/session"
	"github.com/pokt-foundation/portal-api-go/sticky"
)

type RelayResponse struct {
	Err error
}

// TODO: this is needed because pocket-go does not provide an interface yet, which is needed for unit-testing.
// 	remove this and use relayer.Relayer of pocket-go once that is an interface instead of a struct
type pocketRelayer interface {
	Relay(input *relayer.Input, options *provider.RelayRequestOptions) (*relayer.Output, error)
}

type RelayOptions struct {
	// Origin of the request: taken from the http header with the same name
	Origin  string
	Method  string
	RawData string
	Host    string
	// TODO: may need special handling if request are coming from an ALB (application load balancer)
	IP             string
	Path           string
	RequestID      uuid.UUID
	BlockchainID   string
	RpcID          int
	ApplicationID  string
	LoadBalancerID string
}

//TODO: define custom user-errors: e.g. invalid applicationID + error codes should match portal-ai
type Relayer interface {
	RelayWithApp(RelayOptions) error
	RelayWithLb(RelayOptions) error
}

type relayServer struct {
	repository     repository.Repository
	sessionManager session.SessionManager
	nodeSticker    sticky.StickyClientService

	settings RelayerSettings
	relayer  pocketRelayer
	log      *logger.Logger
}

func NewRelayServer(rpcUrls []string, privateKey string, settings RelayerSettings, r repository.Repository, sessionManager session.SessionManager, log *logger.Logger) (Relayer, error) {
	rpcProvider := provider.NewProvider(rpcUrls[0], rpcUrls)
	rpcProvider.UpdateRequestConfig(0, time.Duration(20)*time.Second)

	reqSigner, err := signer.NewSignerFromPrivateKey(privateKey)
	if err != nil {
		log.WithFields(logger.Fields{"error": err}).Warn("Error creating wallet")
		return &relayServer{}, fmt.Errorf("Error creating wallet: %v", err)
	}

	p := relayer.NewRelayer(reqSigner, rpcProvider)

	return &relayServer{
		repository:     r,
		sessionManager: sessionManager,
		relayer:        p,
		settings:       settings,
		log:            log,
	}, nil
}

type relayDetailsBuilder func(repository.Repository, RelayOptions) func(*RelayDetails) error

func appDetails(r repository.Repository, o RelayOptions) func(*RelayDetails) error {
	return func(d *RelayDetails) error {
		app, err := r.GetApplication(o.ApplicationID)
		if err != nil {
			return err
		}
		d.Application = &app
		return nil
	}
}

func chainDetails(r repository.Repository, o RelayOptions) func(*RelayDetails) error {
	return func(d *RelayDetails) error {
		blockchain, err := r.GetBlockchain(o.BlockchainID)
		if err != nil {
			return err
		}
		d.Blockchain = blockchain
		return nil
	}
}

func lbDetails(r repository.Repository, o RelayOptions) func(*RelayDetails) error {
	return func(d *RelayDetails) error {
		lb, err := r.GetLoadBalancer(o.LoadBalancerID)
		if err != nil {
			return err
		}
		d.LoadBalancer = lb
		return nil
	}
}

func detailsBuilder(r repository.Repository, o RelayOptions, builders map[string]relayDetailsBuilder) (*RelayDetails, error) {
	d := RelayDetails{
		RelayOptions: o,
	}
	for k, b := range builders {
		builder := b(r, o)
		if err := builder(&d); err != nil {
			return nil, fmt.Errorf("Error running builder %s: %w", k, err)
		}
	}
	return &d, nil
}

func (r *relayServer) RelayWithApp(relayOptions RelayOptions) error {
	// TODO: metrics recorder

	log := r.log.WithFields(logger.Fields{"relayOptions": relayOptions})

	builders := map[string]relayDetailsBuilder{
		"application": appDetails,
		"blockchain":  chainDetails,
	}
	d, err := detailsBuilder(r.repository, relayOptions, builders)
	if err != nil {
		log.WithFields(logger.Fields{"error": err}).Warn("Error running builder")
		return err
	}
	return r.sendRelay(d)
}

type RelayDetails struct {
	repository.Blockchain
	*repository.Application
	repository.LoadBalancer
	RelayOptions
	sticky.StickyDetails
}

func (r *relayServer) RelayWithLb(relayOptions RelayOptions) error {
	log := r.log.WithFields(logger.Fields{"relayOptions": relayOptions})

	// TODO: verify if order matters here: using maps means no guaranteed order in calling detail builders
	builders := map[string]relayDetailsBuilder{
		"blockchain":   chainDetails,
		"loadbalancer": lbDetails,
	}
	details, err := detailsBuilder(r.repository, relayOptions, builders)
	if err != nil {
		log.WithFields(logger.Fields{"error": err}).Warn("Error running builder")
		return err
	}
	log = log.WithFields(logger.Fields{"RelayDetails": details})

	// TODO: Gigastake Redirect
	if details.LoadBalancer.GigastakeRedirect {
		log.Warn("Gigastake redirect not implemented yet")
		return err
	}

	sd := r.nodeSticker.GetStickyDetails(
		details.LoadBalancer.StickyOptions,
		stickyKeyBuilder(details),
		stickyOptionsVerifier(details.RelayOptions),
	)
	log = log.WithFields(logger.Fields{"stickyDetails": sd})

	selectedApp, err := r.fetchLoadBalancerApplication(details.LoadBalancer, sd.StickyClient.PreferredApplicationID)
	if err != nil {
		// TODO: error code: -32055))
		log.WithFields(logger.Fields{"error": err}).Warn("Error selecting an application for load balancer")
		return err
	}

	details.Application = selectedApp
	if sd.StickyClient.IsEmpty() {
		sd.StickyClient.PreferredApplicationID = selectedApp.ID
	}
	details.StickyDetails = sd
	log.WithFields(logger.Fields{"RelayDetails": details}).Info("Sending relay")

	return r.sendRelay(details)
}

type RelayerSettings struct {
	AatPlan
	DefaultLogLimitBlocks      int
	DefaultStickyOptions       repository.StickyOptions
	DefaultClientStickyOptions sticky.StickyClient
}

// TODO: move to repository package?
type AatPlan string

const (
	AatPlanPremium  AatPlan = "Premium"
	AatPlanFreemium AatPlan = "Freemium"
)

func aatFromApp(app *repository.Application, aatPlan AatPlan) provider.PocketAAT {
	if aatPlan == AatPlanFreemium {
		return provider.PocketAAT{
			Version:      app.GatewayAAT.Version,
			ClientPubKey: app.GatewayAAT.ClientPublicKey,
			AppPubKey:    app.GatewayAAT.ApplicationPublicKey,
			Signature:    app.GatewayAAT.ApplicationSignature,
		}
	}

	return provider.PocketAAT{
		Version:      app.GatewayAAT.Version,
		ClientPubKey: app.GatewayAAT.ClientPublicKey,
		AppPubKey:    app.GatewayAAT.ApplicationPublicKey,
		Signature:    app.GatewayAAT.ApplicationSignature,
	}
}

func FreemiumSettings() RelayerSettings {
	return RelayerSettings{
		AatPlan: AatPlanFreemium,
	}
}

// TODO: return code -32058 in repository package when building LB list (for LBs with no valid app)
func (r *relayServer) fetchLoadBalancerApplication(lb repository.LoadBalancer, preferredApplicationID string) (*repository.Application, error) {
	// TODO: add a service that maintains verified Application IDs for a LB: it invalidates and reloads every set interval

	apps := lb.Applications
	if len(apps) < 1 {
		// TODO: return error code -32058
		return &repository.Application{}, fmt.Errorf("Load Balancer %s configuration invalid: no valid applications", lb.ID)
	}

	// TODO: remove once LBs applications are returned in a map[AppID]*Application
	for _, app := range apps {
		if app.ID == preferredApplicationID {
			return app, nil
		}
	}

	return apps[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(apps))], nil
}

func (r *relayServer) sendRelay(details *RelayDetails) error {

	log := r.log.WithFields(logger.Fields{"relayDetails": details})

	// TODO: validate request:
	// secretKeyValidator, // checkSecretKey(application, secretKeyDetails)
	// whilelistValidator, // whitelistValidator: 1.origins (err code: -32060), 2.userAgents: (err code: -32061)

	pocketAat := aatFromApp(details.Application, r.settings.AatPlan)
	log = log.WithFields(logger.Fields{"pocketAAT": pocketAat})

	session, err := r.sessionManager.GetSession(session.Key{PublicKey: pocketAat.AppPubKey, BlockchainID: details.Blockchain.ID})
	if err != nil {
		log.WithFields(logger.Fields{"error": err}).Warn("Error getting session")
		return err
	}
	startTime := time.Now()
	log = log.WithFields(logger.Fields{"session": session, "startTime": startTime})

	// TODO: metrics recorder

	// Node selection:
	// Get session's available (i.e. not exhausted) nodes
	// ChainCheck + SyncChecka (from fisherman if possible)
	// StickyNodes
	// CherryPicker
	// EVM/non-EVM restrictions
	// -------------------
	var node *provider.Node
	node = nodeFromAddress(details.StickyDetails.StickyClient.PreferredNodeAddress, session)
	if node == nil {
		node = firstNode(session)
	}
	if node == nil {
		log.Warn("Session has no nodes")
		return fmt.Errorf("Session has no nodes")
	}

	// TODO: going down multiple layers usually indicates a design issue: can this be improved?
	if details.StickyDetails.StickyClient.IsEmpty() {
		details.StickyDetails.StickyClient = sticky.StickyClient{
			PreferredApplicationID: details.Application.ID,
			PreferredNodeAddress:   node.Address,
		}
	}

	log = log.WithFields(logger.Fields{"node": node})
	log.Info("Sending relay")

	relay := relayer.Input{
		Method:     "POST",
		PocketAAT:  &pocketAat,
		Session:    session,
		Node:       node,
		Blockchain: details.Blockchain.ID,
		Data:       details.RelayOptions.RawData,
		Path:       details.RelayOptions.Path,
	}

	relayOutput, err := r.relayer.Relay(&relay, nil)
	if err != nil {
		log.WithFields(logger.Fields{"error": err}).Info("Error relaying")
		// TODO: differentiate user errors from node errors
		err := r.nodeSticker.Failure(&details.StickyDetails)
		return err
	}

	r.nodeSticker.Success(&details.StickyDetails)
	log.WithFields(logger.Fields{"relayOutput": relayOutput}).Info("Received relay response")

	return parseRelayResponse(relayOutput)
}

func nodeFromAddress(address string, session *provider.Session) *provider.Node {
	for _, n := range session.Nodes {
		if n.Address == address {
			return n
		}
	}
	return nil
}

func firstNode(session *provider.Session) *provider.Node {
	if session == nil || len(session.Nodes) == 0 {
		return nil
	}
	return session.Nodes[0]
}

// TODO: This likely belongs in pocket-go: a function that can process the Output struct returned by pocket-go/relayer
func parseRelayResponse(r *relayer.Output) error {
	return nil
}

func stickyKeyBuilder(d *RelayDetails) sticky.KeyBuilder {
	return func(o repository.StickyOptions) sticky.Key {
		k := sticky.Key{
			IP:           d.RelayOptions.IP,
			BlockchainID: d.Blockchain.ID,
		}

		k.ApplicationID = d.Application.ID
		k.LoadBalancerID = d.LoadBalancer.ID
		return k
	}
}

func stickyOptionsVerifier(r RelayOptions) sticky.OptionsVerifier {
	// Users/bots could fetch several origins from the same ip which not all allow stickiness,
	// this is needed to not trigger stickiness on those other origins if is already saved.
	return func(o repository.StickyOptions) error {
		// TODO: Can this be a function of StickyOpts struct?
		originLowerCase := strings.ToLower(r.Origin)
		for _, stickyOrigin := range o.StickyOrigins {
			if strings.Contains(originLowerCase, strings.ToLower(stickyOrigin)) {
				return nil
			}
		}
		return fmt.Errorf("Origin %q does not match sticky origins", r.Origin)
	}
}
