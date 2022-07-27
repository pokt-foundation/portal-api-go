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

type Blockchain struct {
	ID                string   `json:"id"`
	Altruist          string   `json:"altruist"`
	Blockchain        string   `json:"blockchain"`
	BlockchainAliases []string `json:"blockchainAliase"`
	ChainID           string   `json:"chainID"`
	ChaindIDCheck     string   `json:"chainIDCheck"`
	Description       string   `json:"description"`
	Index             int64    `json:"index"`
	LogLimitBlocks    int64    `json:"logLimitBlocks"`
	Network           string   `json:"network"`
	NetworkID         string   `json:"networkID"`
	Path              string   `json:"path"`
	RequestTimeout    int64    `json:"requestTimeout"`
	Ticker            string   `json:"ticker"`
}

type User struct {
	ID string `json:"id"`
}

// LBs also contain this struct
type StickyOptions struct {
	Stickiness     bool
	Duration       string
	UseRPCID       bool
	RelaysLimit    int
	StickyOrigins  []string
	RpcIDThreshold int
}

func (s *StickyOptions) IsEmpty() bool {
	if !s.Stickiness {
		return true
	}
	return len(s.StickyOrigins) == 0
}

// loadBalancer is an internal struct, reflects json, contains unverified fields, e.g. applicationIDs
type loadBalancer struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ApplicationIDs []string `json:"applicationIDs"`
	// User []*User
	// TODO: load from db table/view
	StickyOptions
}

// LoadBalancer contains verified fields, e.g. applications
type LoadBalancer struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	RequestTimeout int64     `json:"requestTimeout"`
	Gigastake      bool      `json:"gigastake"`
	UserID         string    `json:"userID"`
	ApplicationIDs []string  `json:"applicationIDs,omitempty"`
	// TODO: use map[AppID]*Application instead (to speed-up fetchLoadBalancerApplication routine)
	Applications []*Application
	// User []*User
	StickyOptions
	// TODO: this likely needs to be replaced with gigastake apps
	GigastakeRedirect bool
}

// TODO: identify fields that should be stored encrypted in-memory
type Application struct {
	ID                         string     `json:"id"`
	ContactEmail               string     `json:"contactEmail"`
	CreatedAt                  *time.Time `json:"createdAt"`
	Description                string     `json:"description"`
	Name                       string     `json:"name"`
	Owner                      string     `json:"owner"`
	UpdatedAt                  *time.Time `json:"updatedAt"`
	URL                        string     `json:"url"`
	UserID                     string     `json:"userID"`
	PublicPocketAccount        `json:"publicPocketAccount"`
	FreeTierApplicationAccount `json:"freeTierApplicationAccount"`
	FreeTierAAT                `json:"freeTierAAT"`
	GatewayAAT                 `json:"gatewatAAT"`
	GatewaySettings            `json:"gatewaySettings"`
}

type PublicPocketAccount struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
}

type FreeTierApplicationAccount struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	// TODO: likely need to store an encrypted form in memory
	PrivateKey string `json:"privateKey"`
	Version    string `json:"version"`
}

type GatewayAAT struct {
	Version              string `json:"version"`
	ApplicationPublicKey string `json:"applicationPublicKey"`
	ClientPublicKey      string `json:"clientPublicKey"`
	ApplicationSignature string `json:"applicationSignature"`
}

type FreeTierAAT struct {
	Version              string `json:"version"`
	ClientPublicKey      string `json:"clientPublicKey"`
	ApplicationPublicKey string `json:"applicationPublicKey"`
	ApplicationSignature string `json:"applicationSignature"`
}

type GatewaySettings struct {
	SecretKey           string   `json:"secretKey"`
	SecretKeyRequired   bool     `json:"secreyKeyRequired"`
	WhitelistOrigins    []string `json:"whitelistOrigins,omitempty"`
	WhitelistUserAgents []string `json:"whitelistUserAgents,omitempty"`
	// TODO: change next two whitelist to use their structs
	WhitelistContracts   []string `json:"whitelistContracts,omitempty"`
	WhitelistMethods     []string `json:"whitelistMethods,omitempty"`
	WhitelistBlockchains []string `json:"whitelistBlockchains,omitempty"`
}

type WhitelistContract struct {
	BlockchainID string   `json:"blockchainID"`
	Contracts    []string `json:"contracts"`
}

type WhitelistMethod struct {
	BlockchainID string
	Methods      []string `json:"methods"`
}

type Repository interface {
	GetApplication(id string) (Application, error)
	GetBlockchain(alias string) (Blockchain, error)
	GetLoadBalancer(id string) (LoadBalancer, error)
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

type AppStickiness struct {
	ApplicationID string
	NodeAddress   string
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
