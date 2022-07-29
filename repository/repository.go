package repository

import (
	"encoding/json"
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

// TODO: identify fields that should be stored encrypted in-memory
type Application struct {
	ID                         string                     `json:"id"`
	UserID                     string                     `json:"userID"`
	Name                       string                     `json:"name"`
	Status                     string                     `json:"status"`
	ContactEmail               string                     `json:"contactEmail"`
	Description                string                     `json:"description"`
	Owner                      string                     `json:"owner"`
	URL                        string                     `json:"url"`
<<<<<<< HEAD
	Dummy                      bool                       `json:"dummy"`
	MaxRelays                  int                        `json:"maxRelays"`
=======
	MaxRelays                  int64                      `json:"maxRelays"`
	Dummy                      bool                       `json:"dummy"`
>>>>>>> 77af657 (feat: added the fields in the PUB structs to the repository and postgresdriver packages.)
	FreeTier                   bool                       `json:"freeTier"`
	FreeTierAAT                FreeTierAAT                `json:"freeTierAAT"`
	FreeTierApplicationAccount FreeTierApplicationAccount `json:"freeTierApplicationAccount"`
	GatewayAAT                 GatewayAAT                 `json:"gatewatAAT"`
	GatewaySettings            GatewaySettings            `json:"gatewaySettings"`
	NotificationSettings       NotificationSettings       `json:"notificationSettings"`
	PublicPocketAccount        PublicPocketAccount        `json:"publicPocketAccount"`
<<<<<<< HEAD
	CreatedAt                  time.Time                  `json:"createdAt"`
	UpdatedAt                  time.Time                  `json:"updatedAt"`
=======
	CreatedAt                  *time.Time                 `json:"createdAt"`
	UpdatedAt                  *time.Time                 `json:"updatedAt"`
>>>>>>> 77af657 (feat: added the fields in the PUB structs to the repository and postgresdriver packages.)
}

type FreeTierAAT struct {
	Version              string `json:"version"`
	ClientPublicKey      string `json:"clientPublicKey"`
	ApplicationPublicKey string `json:"applicationPublicKey"`
	ApplicationSignature string `json:"applicationSignature"`
}

type FreeTierApplicationAccount struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	// TODO: likely need to store an encrypted form in memory
	PrivateKey string `json:"privateKey"`
	Version    string `json:"version"`
}

type GatewayAAT struct {
	ApplicationPublicKey string `json:"applicationPublicKey"`
	ApplicationSignature string `json:"applicationSignature"`
	ClientPublicKey      string `json:"clientPublicKey"`
	Version              string `json:"version"`
}

type GatewaySettings struct {
	SecretKey            string              `json:"secretKey"`
	SecretKeyRequired    bool                `json:"secreyKeyRequired"`
	WhitelistOrigins     []string            `json:"whitelistOrigins,omitempty"`
	WhitelistUserAgents  []string            `json:"whitelistUserAgents,omitempty"`
	WhitelistContracts   []WhitelistContract `json:"whitelistContracts,omitempty"`
	WhitelistMethods     []WhitelistMethod   `json:"whitelistMethods,omitempty"`
	WhitelistBlockchains []string            `json:"whitelistBlockchains,omitempty"`
}

type WhitelistContract struct {
	BlockchainID string   `json:"blockchainID"`
	Contracts    []string `json:"contracts"`
}

type WhitelistMethod struct {
	BlockchainID string
	Methods      []string `json:"methods"`
}

type NotificationSettings struct {
	SignedUp      bool `json:"signedUp"`
	Quarter       bool `json:"quarter"`
	Half          bool `json:"half"`
	ThreeQuarters bool `json:"threeQuarters"`
	Full          bool `json:"full"`
}

type PublicPocketAccount struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
}

type Blockchain struct {
	ID                string           `json:"id"`
	Altruist          string           `json:"altruist"`
	Blockchain        string           `json:"blockchain"`
	ChainID           string           `json:"chainID"`
<<<<<<< HEAD
	ChainIDCheck      string           `json:"chainIDCheck"`
=======
	ChaindIDCheck     string           `json:"chainIDCheck"`
>>>>>>> 77af657 (feat: added the fields in the PUB structs to the repository and postgresdriver packages.)
	Description       string           `json:"description"`
	EnforceResult     string           `json:"enforceResult"`
	Network           string           `json:"network"`
	NetworkID         string           `json:"networkID"`
	Path              string           `json:"path"`
	SyncCheck         string           `json:"syncCheck"`
	Ticker            string           `json:"ticker"`
<<<<<<< HEAD
	BlockchainAliases []string         `json:"blockchainAliases"`
	RequestTimeout    int              `json:"requestTimeout"`
	Index             int              `json:"index"`
	LogLimitBlocks    int              `json:"logLimitBlocks"`
	SyncAllowance     int              `json:"syncAllowance"`
=======
	BlockchainAliases []string         `json:"blockchainAliase"`
	Index             int64            `json:"index"`
	LogLimitBlocks    int64            `json:"logLimitBlocks"`
	RequestTimeout    int64            `json:"requestTimeout"`
	SyncAllowance     int64            `json:"syncAllowance"`
>>>>>>> 77af657 (feat: added the fields in the PUB structs to the repository and postgresdriver packages.)
	Active            bool             `json:"active"`
	Redirects         []Redirects      `json:"redirects"`
	SyncCheckOptions  SyncCheckOptions `json:"syncCheckOptions"`
}

// TODO - Figure out how to handle Redirects field in Postgres
type Redirects struct {
	Alias          string `json:"alias"`
	Domain         string `json:"domain"`
	LoadBalancerID string `json:"loadBalancerID"`
}

type SyncCheckOptions struct {
	Body      string `json:"body"`
	ResultKey string `json:"resultKey"`
	Path      string `json:"path"`
<<<<<<< HEAD
	Allowance int    `json:"allowance"`
=======
	Allowance int64  `json:"allowance"`
>>>>>>> 77af657 (feat: added the fields in the PUB structs to the repository and postgresdriver packages.)
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

<<<<<<< HEAD
// LoadBalancer contains verified fields, e.g. applications (referred to as Endpoints on the Portal UI frontend/API)
=======
// LoadBalancer contains verified fields, e.g. applications (Referred to as Endpoint in Portal UI Backend)
>>>>>>> 77af657 (feat: added the fields in the PUB structs to the repository and postgresdriver packages.)
type LoadBalancer struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	UserID         string   `json:"userID"`
	ApplicationIDs []string `json:"applicationIDs,omitempty"`
<<<<<<< HEAD
	RequestTimeout int      `json:"requestTimeout"`
=======
	RequestTimeout int64    `json:"requestTimeout"`
>>>>>>> 77af657 (feat: added the fields in the PUB structs to the repository and postgresdriver packages.)
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

type StickyOptions struct {
	Duration       string   `json:"duration"`
	StickyOrigins  []string `json:"stickyOrigins"`
	RelaysLimit    int      `json:"relaysLimit"`
	RpcIDThreshold int
	Stickiness     bool `json:"stickiness"`
	StickinessTemp bool `json:"stickinessTemp"`
	UseRPCID       bool `json:"useRPCID"`
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
