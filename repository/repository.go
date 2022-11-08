package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"
)

type Repository interface {
	GetApplication(id string) (Application, error)
	GetBlockchain(alias string) (Blockchain, error)
	GetLoadBalancer(id string) (LoadBalancer, error)
}

var (
	ErrNoFieldsToUpdate               = errors.New("no fields to update")
	ErrInvalidAppStatus               = errors.New("invalid app status")
	ErrInvalidPayPlanType             = errors.New("invalid pay plan type")
	ErrNotEnterprisePlan              = errors.New("custom limits may only be set on enterprise plans")
	ErrEnterprisePlanNeedsCustomLimit = errors.New("enterprise plans must have a custom limit set")
)

// TODO: identify fields that should be stored encrypted in-memory
type Application struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"userID"`
	Name               string    `json:"name"`
	ContactEmail       string    `json:"contactEmail"`
	Description        string    `json:"description"`
	Owner              string    `json:"owner"`
	URL                string    `json:"url"`
	Dummy              bool      `json:"dummy"`
	Status             AppStatus `json:"status"`
	FirstDateSurpassed time.Time `json:"firstDateSurpassed"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`

	GatewayAAT           GatewayAAT           `json:"gatewayAAT"`
	GatewaySettings      GatewaySettings      `json:"gatewaySettings"`
	Limit                AppLimit             `json:"limit"`
	NotificationSettings NotificationSettings `json:"notificationSettings"`
}

func (a *Application) Table() Table {
	return TableApplications
}

func (a *Application) DailyLimit() int {
	if a.Limit.PayPlan.Type == Enterprise {
		return a.Limit.CustomLimit
	}

	return a.Limit.PayPlan.Limit
}

func (a *Application) IsInvalid() error {
	if !ValidAppStatuses[a.Status] {
		return ErrInvalidAppStatus
	}

	if !ValidPayPlanTypes[a.Limit.PayPlan.Type] {
		return ErrInvalidPayPlanType
	}

	if a.Limit.PayPlan.Type != Enterprise && a.Limit.CustomLimit != 0 {
		return ErrNotEnterprisePlan
	}
	return nil
}

type AppLimit struct {
	ID          string  `json:"id,omitempty"`
	PayPlan     PayPlan `json:"payPlan"`
	CustomLimit int     `json:"customLimit"`
}

func (a *AppLimit) Table() Table {
	return TableAppLimits
}

type AppStatus string

const (
	AwaitingFreetierFunds   AppStatus = "AWAITING_FREETIER_FUNDS"
	AwaitingFreetierStaking AppStatus = "AWAITING_FREETIER_STAKING"
	AwaitingFunds           AppStatus = "AWAITING_FUNDS"
	AwaitingFundsRemoval    AppStatus = "AWAITING_FUNDS_REMOVAL"
	AwaitingGracePeriod     AppStatus = "AWAITING_GRACE_PERIOD"
	AwaitingSlotFunds       AppStatus = "AWAITING_SLOT_FUNDS"
	AwaitingSlotStaking     AppStatus = "AWAITING_SLOT_STAKING"
	AwaitingStaking         AppStatus = "AWAITING_STAKING"
	AwaitingUnstaking       AppStatus = "AWAITING_UNSTAKING"
	Decomissioned           AppStatus = "DECOMISSIONED"
	InService               AppStatus = "IN_SERVICE"
	Orphaned                AppStatus = "ORPHANED"
	Ready                   AppStatus = "READY"
	Swappable               AppStatus = "SWAPPABLE"
)

var (
	ValidAppStatuses = map[AppStatus]bool{
		"":                      true, // needed since it can be empty too
		AwaitingFreetierFunds:   true,
		AwaitingFreetierStaking: true,
		AwaitingFunds:           true,
		AwaitingFundsRemoval:    true,
		AwaitingGracePeriod:     true,
		AwaitingSlotFunds:       true,
		AwaitingSlotStaking:     true,
		AwaitingStaking:         true,
		AwaitingUnstaking:       true,
		Decomissioned:           true,
		InService:               true,
		Orphaned:                true,
		Ready:                   true,
		Swappable:               true,
	}
)

type PayPlan struct {
	Type  PayPlanType `json:"planType"`
	Limit int         `json:"dailyLimit"`
}

type PayPlanType string

const (
	TestPlanV0   PayPlanType = "TEST_PLAN_V0"
	TestPlan10K  PayPlanType = "TEST_PLAN_10K"
	TestPlan90k  PayPlanType = "TEST_PLAN_90K"
	FreetierV0   PayPlanType = "FREETIER_V0"
	PayAsYouGoV0 PayPlanType = "PAY_AS_YOU_GO_V0"
	Enterprise   PayPlanType = "ENTERPRISE"
)

var (
	ValidPayPlanTypes = map[PayPlanType]bool{
		"":           true, // needs to be allowed while the change for all apps to have plans is done
		TestPlanV0:   true,
		TestPlan10K:  true,
		TestPlan90k:  true,
		FreetierV0:   true,
		PayAsYouGoV0: true,
		Enterprise:   true,
	}
)

type AppLimits struct {
	AppID                string                `json:"appID,omitempty"`
	AppName              string                `json:"appName,omitempty"`
	AppUserID            string                `json:"appUserID,omitempty"`
	PublicKey            string                `json:"publicKey,omitempty"`
	PlanType             PayPlanType           `json:"planType"`
	DailyLimit           int                   `json:"dailyLimit"`
	FirstDateSurpassed   *time.Time            `json:"firstDateSurpassed,omitempty"`
	NotificationSettings *NotificationSettings `json:"notificationSettings,omitempty"`
}

// UpdateApplication struct holding possible fields to update
type UpdateApplication struct {
	Name                 string                `json:"name,omitempty"`
	Status               AppStatus             `json:"status,omitempty"`
	FirstDateSurpassed   time.Time             `json:"firstDateSurpassed,omitempty"`
	GatewaySettings      *GatewaySettings      `json:"gatewaySettings,omitempty"`
	NotificationSettings *NotificationSettings `json:"notificationSettings,omitempty"`
	Limit                *AppLimit             `json:"appLimit,omitempty"`
	Remove               bool                  `json:"remove,omitempty"`
}

func (u *UpdateApplication) IsInvalid() error {
	if u == nil {
		return ErrNoFieldsToUpdate
	}
	if !ValidAppStatuses[u.Status] {
		return ErrInvalidAppStatus
	}
	if u.Limit != nil && !ValidPayPlanTypes[u.Limit.PayPlan.Type] {
		return ErrInvalidPayPlanType
	}
	if u.Limit != nil && u.Limit.PayPlan.Type != Enterprise && u.Limit.CustomLimit != 0 {
		return ErrNotEnterprisePlan
	}
	if u.Limit != nil && u.Limit.PayPlan.Type == Enterprise && u.Limit.CustomLimit == 0 {
		return ErrEnterprisePlanNeedsCustomLimit
	}
	return nil
}

type UpdateFirstDateSurpassed struct {
	ApplicationIDs     []string  `json:"applicationIDs"`
	FirstDateSurpassed time.Time `json:"firstDateSurpassed"`
}

type GatewayAAT struct {
	ID                   string `json:"id,omitempty"`
	Address              string `json:"address"`
	ApplicationPublicKey string `json:"applicationPublicKey"`
	ApplicationSignature string `json:"applicationSignature"`
	ClientPublicKey      string `json:"clientPublicKey"`
	PrivateKey           string `json:"privateKey"`
	Version              string `json:"version"`
}

func (a *GatewayAAT) Table() Table {
	return TableGatewayAAT
}

type GatewaySettings struct {
	ID                   string              `json:"id,omitempty"`
	SecretKey            string              `json:"secretKey"`
	SecretKeyRequired    bool                `json:"secretKeyRequired"`
	WhitelistOrigins     []string            `json:"whitelistOrigins,omitempty"`
	WhitelistUserAgents  []string            `json:"whitelistUserAgents,omitempty"`
	WhitelistContracts   []WhitelistContract `json:"whitelistContracts,omitempty"`
	WhitelistMethods     []WhitelistMethod   `json:"whitelistMethods,omitempty"`
	WhitelistBlockchains []string            `json:"whitelistBlockchains,omitempty"`
}

func (s *GatewaySettings) Table() Table {
	return TableGatewaySettings
}

type WhitelistContract struct {
	BlockchainID string   `json:"blockchainID"`
	Contracts    []string `json:"contracts"`
}

type WhitelistMethod struct {
	BlockchainID string   `json:"blockchainID"`
	Methods      []string `json:"methods"`
}

type NotificationSettings struct {
	ID            string `json:"id,omitempty"`
	SignedUp      bool   `json:"signedUp"`
	Quarter       bool   `json:"quarter"`
	Half          bool   `json:"half"`
	ThreeQuarters bool   `json:"threeQuarters"`
	Full          bool   `json:"full"`
}

func (s *NotificationSettings) Table() Table {
	return TableNotificationSettings
}

type Blockchain struct {
	ID                string           `json:"id"`
	Altruist          string           `json:"altruist"`
	Blockchain        string           `json:"blockchain"`
	ChainID           string           `json:"chainID"`
	ChainIDCheck      string           `json:"chainIDCheck"`
	Description       string           `json:"description"`
	EnforceResult     string           `json:"enforceResult"`
	Network           string           `json:"network"`
	Path              string           `json:"path"`
	SyncCheck         string           `json:"syncCheck"`
	Ticker            string           `json:"ticker"`
	BlockchainAliases []string         `json:"blockchainAliases"`
	LogLimitBlocks    int              `json:"logLimitBlocks"`
	RequestTimeout    int              `json:"requestTimeout"`
	SyncAllowance     int              `json:"syncAllowance"`
	Active            bool             `json:"active"`
	Redirects         []Redirect       `json:"redirects"`
	SyncCheckOptions  SyncCheckOptions `json:"syncCheckOptions"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
}

func (b *Blockchain) Table() Table {
	return TableBlockchains
}

type Redirect struct {
	ID             string    `json:"id"`
	BlockchainID   string    `json:"blockchainID"`
	Alias          string    `json:"alias"`
	Domain         string    `json:"domain"`
	LoadBalancerID string    `json:"loadBalancerID"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (r *Redirect) Table() Table {
	return TableRedirects
}

type SyncCheckOptions struct {
	BlockchainID string `json:"blockchainID"`
	Body         string `json:"body"`
	Path         string `json:"path"`
	ResultKey    string `json:"resultKey"`
	Allowance    int    `json:"allowance"`
}

func (o *SyncCheckOptions) Table() Table {
	return TableSyncCheckOptions
}

// loadBalancer is an internal struct, reflects json, contains unverified fields, e.g. applicationIDs
type loadBalancer struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ApplicationIDs []string `json:"applicationIDs"`
	// User []*User
	// TODO: load from db table/view
	StickyOptions StickyOptions `json:"stickinessOptions"`
}

// LoadBalancer contains verified fields, e.g. applications (referred to as Endpoints on the Portal UI frontend/API)
type LoadBalancer struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	UserID         string   `json:"userID"`
	ApplicationIDs []string `json:"applicationIDs,omitempty"`
	RequestTimeout int      `json:"requestTimeout"`
	Gigastake      bool     `json:"gigastake"`
	// TODO: this likely needs to be replaced with gigastake apps
	GigastakeRedirect bool          `json:"gigastakeRedirect"`
	StickyOptions     StickyOptions `json:"stickinessOptions"`
	// TODO: use map[AppID]*Application instead (to speed-up fetchLoadBalancerApplication routine)
	Applications []*Application
	// User []*User
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (l *LoadBalancer) Table() Table {
	return TableLoadBalancers
}

// LbApp represents in DB relationships of lb and apps
// do not change the tags, they're snake_case on purpose
type LbApp struct {
	LbID  string `json:"lb_id"`
	AppID string `json:"app_id"`
}

func (l *LbApp) Table() Table {
	return TableLbApps
}

// UpdateLoadBalancer struct holding possible field to update
type UpdateLoadBalancer struct {
	Name          string         `json:"name,omitempty"`
	StickyOptions *StickyOptions `json:"stickinessOptions,omitempty"`
	Remove        bool           `json:"remove,omitempty"`
}

type StickyOptions struct {
	ID            string   `json:"id,omitempty"`
	Duration      string   `json:"duration"`
	StickyOrigins []string `json:"stickyOrigins"`
	StickyMax     int      `json:"stickyMax"`
	Stickiness    bool     `json:"stickiness"`
}

func (s *StickyOptions) Table() Table {
	return TableStickinessOptions
}

func (s *StickyOptions) IsEmpty() bool {
	if !s.Stickiness {
		return true
	}
	return len(s.StickyOrigins) == 0
}

type User struct {
	ID string `json:"id"`
}

type AppStickiness struct {
	ApplicationID string
	NodeAddress   string
}

func NewRepository(jsonFilesPath string, log *logger.Logger) (Repository, error) {
	l := log.WithFields(logger.Fields{"path": jsonFilesPath})

	// TODO: use a db (postgres?) backend once migration from mongo is done.
	//	for now, just load from json files
	blockchains, err := loadBlockchains(path.Join(jsonFilesPath, "Blockchains.json"))
	if err != nil {
		return nil, fmt.Errorf("Error loading blockchains: %v", err)
	}

	applications, err := loadApplications(path.Join(jsonFilesPath, "Applications.json"))
	if err != nil {
		return nil, fmt.Errorf("Error loading applications: %v", err)
	}

	allLbs, err := loadLoadBalancers(path.Join(jsonFilesPath, "LoadBalancers.json"))
	if err != nil {
		return nil, fmt.Errorf("Error loading load balancers: %v", err)
	}

	lbs, invalidAppIDs, err := buildLoadBalancers(allLbs, applications)
	if err != nil {
		return nil, fmt.Errorf("Error loading load balancers: %v", err)
	}
	if len(invalidAppIDs) > 0 {
		l.WithFields(logger.Fields{"invalidApplicationIDs": invalidAppIDs}).Warnf("One or more of the specified application IDs were invalid.")
	}

	return &cachingRepository{
		blockchains:   blockchains,
		apps:          applications,
		loadbalancers: lbs,
		log:           log,
	}, nil
}

// TODO: caching
type cachingRepository struct {
	apps map[string]Application
	// TODO: if it helps performance, use a map: multiple keys (i.e. aliases of a blockchain) can refer to the same value (i.e. address of Blockchain struct)
	blockchains   []Blockchain
	loadbalancers map[string]LoadBalancer

	log *logger.Logger
}

func (c *cachingRepository) GetApplication(id string) (Application, error) {
	if app, ok := c.apps[id]; ok {
		return app, nil
	}
	return Application{}, fmt.Errorf("No applications found matching %s", id)
}

// TODO: are these replacements needed?
// syncCheckOptions.body -> strings.ReplaceAll(body, `\\"`, `"`)
// chainIDCheck -> strings.Replace(All?)(chainIDCheck, `\\"`, `"`)
func (c *cachingRepository) GetBlockchain(alias string) (Blockchain, error) {
	// TODO: throw error -32057 on blockchain not found
	return blockchainForAlias(alias, c.blockchains)
}

func (c *cachingRepository) GetLoadBalancer(id string) (LoadBalancer, error) {
	if lb, ok := c.loadbalancers[id]; ok {
		return lb, nil
	}

	return LoadBalancer{}, fmt.Errorf("No loadbalancers found matching %s", id)
}

func blockchainForAlias(alias string, blockchains []Blockchain) (Blockchain, error) {
	lowercaseAlias := strings.ToLower(alias)
	for _, b := range blockchains {
		for _, alias := range b.BlockchainAliases {
			if strings.ToLower(alias) == lowercaseAlias {
				return b, nil
			}
		}
	}
	return Blockchain{}, fmt.Errorf("No blockchains found matching %s", lowercaseAlias)
}

// TODO: these can be moved to the repository_test.go file once we have integrated with a db.
//	They will still be needed for testing purposes, loading a subset of data from json-formatted files.
func loadBlockchains(file string) ([]Blockchain, error) {
	var blockchains []Blockchain
	err := loadData(file, &blockchains)
	return blockchains, err
}

// TODO: apps should use a map[string]Application
func loadApplications(file string) (map[string]Application, error) {
	var applications []Application
	err := loadData(file, &applications)
	if err != nil {
		return nil, err
	}

	m := make(map[string]Application)
	for _, a := range applications {
		m[a.ID] = a
	}
	return m, nil
}

func loadLoadBalancers(file string) ([]loadBalancer, error) {
	var lbs []loadBalancer
	err := loadData(file, &lbs)
	return lbs, err
}

// Returns invalid application IDs as the second return value.
// TODO: remove this check once postgres migration is done as db will check for data integrity
func buildLoadBalancers(items []loadBalancer, apps map[string]Application) (map[string]LoadBalancer, map[string][]string, error) {
	lbs := make(map[string]LoadBalancer)
	invalid := make(map[string][]string)
	for _, lb := range items {
		var verifiedApps []*Application
		for _, id := range lb.ApplicationIDs {
			if app, ok := apps[id]; ok {
				verifiedApps = append(verifiedApps, &app)
			} else {
				invalid[lb.ID] = append(invalid[lb.ID], id)
			}
		}
		lbs[lb.ID] = LoadBalancer{
			ID:           lb.ID,
			Name:         lb.Name,
			Applications: verifiedApps,
		}
	}
	return lbs, invalid, nil
}

func loadData(file string, data interface{}) error {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(contents, data)
}

type Table string

const (
	TableLoadBalancers        Table = "loadbalancers"
	TableStickinessOptions    Table = "stickiness_options"
	TableLbApps               Table = "lb_apps"
	TableApplications         Table = "applications"
	TableAppLimits            Table = "app_limits"
	TableGatewayAAT           Table = "gateway_aat"
	TableGatewaySettings      Table = "gateway_settings"
	TableNotificationSettings Table = "notification_settings"
	TableBlockchains          Table = "blockchains"
	TableRedirects            Table = "redirects"
	TableSyncCheckOptions     Table = "sync_check_options"
)

type Action string

const (
	ActionInsert Action = "INSERT"
	ActionUpdate Action = "UPDATE"
)

type Notification struct {
	Table  Table
	Action Action
	Data   SavedOnDB
}

type SavedOnDB interface {
	Table() Table
}
