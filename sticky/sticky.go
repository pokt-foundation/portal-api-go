package sticky

import (
	"fmt"
	"time"

	logger "github.com/sirupsen/logrus"

	"github.com/pokt-foundation/portal-api-go/repository"
)

// sticky nodes responsibilities (so far):
//	- Maintain the valid stickyClient structs
//		- Add/Remove items based on Relay results
//		- Maintain error counts and relay counts for the specified TTL
//		- Remove items based on relays/errors limit/counts.

type StickyClientService interface {
	GetStickyDetails(repository.StickyOptions, KeyBuilder, OptionsVerifier) StickyDetails
	Success(*StickyDetails) error
	Failure(*StickyDetails) error
}

type Key struct {
	RPCID          int
	LoadBalancerID string
	ApplicationID  string
	BlockchainID   string
	IP             string
}

type KeyBuilder func(repository.StickyOptions) Key
type OptionsVerifier func(repository.StickyOptions) error

// Note: PreferredNode, relayLimit and errorCount have separate TTLs
type StickyClient struct {
	PreferredApplicationID string
	PreferredNodeAddress   string
	RPCID                  int

	Relays *CountWithTTL
	Errors *CountWithTTL
}

func (c StickyClient) IsEmpty() bool {
	return c.PreferredApplicationID == "" && c.PreferredNodeAddress == ""
}

type StickyNodeSettings struct {
	Duration                 time.Duration
	RelayLimit               int
	MaxErrors                int
	DefaultStickinessOptions repository.StickyOptions
	DefaultStickyClient      StickyClient
}

func NewStickyNodes(settings StickyNodeSettings) StickyClientService {
	return &stickyNodes{
		settings: settings,
		items:    make(map[Key]StickyClient),
	}
}

// TODO: Add a RW mutex
type stickyNodes struct {
	settings StickyNodeSettings
	items    map[Key]StickyClient

	log *logger.Logger
}

func (s *stickyNodes) Get(k Key) StickyClient {
	return s.validate(k)
}

// validate performs the following checks before returning a StickyClient:
// TTL of ErrorCount
// TTL of RelayLimit
func (s *stickyNodes) validate(k Key) StickyClient {
	sc, ok := s.items[k]
	if !ok {
		return StickyClient{}
	}

	now := time.Now()
	if sc.Errors.IsExpired(now) || sc.Relays.IsExpired(now) {
		s.items[k] = sc
	}
	return sc
}

// StickyDetails is used to pass around a Sticky item with its corresponding key.
//	This is to avoid having to store the key in the sticky cache
type StickyDetails struct {
	Key
	repository.StickyOptions
	StickyClient
}

// TODO: once implementation is done, re-evaluate whether this function can be combined for an easier interface to stickyClientService
// Replaces checkClientStickiness
// TODO: find better names for StickyDetails function (and StickyClient struct)
func (s *stickyNodes) GetStickyDetails(o repository.StickyOptions, keyBuilder KeyBuilder, verifier OptionsVerifier) StickyDetails {

	stickyOpts := s.settings.DefaultStickinessOptions
	if !o.IsEmpty() {
		stickyOpts = o
	}
	log := s.log.WithFields(logger.Fields{"stickyOptions": stickyOpts})

	if !stickyOpts.Stickiness {
		return StickyDetails{StickyClient: s.settings.DefaultStickyClient}
	}

	key := keyBuilder(stickyOpts)
	if key.IsEmpty() {
		log.WithFields(logger.Fields{"Key": key}).Info("Empty sticky look-up key")
		return StickyDetails{}
	}

	if err := verifier(stickyOpts); err != nil {
		log.WithFields(logger.Fields{"error": err}).Info("sticky options verification failed")
		return StickyDetails{}
	}

	return StickyDetails{
		StickyOptions: stickyOpts,
		StickyClient:  s.Get(key),
		Key:           key,
	}
}

// TODO: move defaultLogLimitBlocks to settings of relayServer
func (k Key) IsEmpty() bool {
	return k.LoadBalancerID == "" && k.ApplicationID == "" && k.RPCID <= 0
}

func (k Key) String() string {
	prefix := k.ApplicationID
	if prefix == "" {
		prefix = k.LoadBalancerID
	}

	if prefix == "" {
		return fmt.Sprintf("%d-%s-%s", k.RPCID, k.IP, k.BlockchainID)
	}
	return fmt.Sprintf("%s-%s-%s", prefix, k.IP, k.BlockchainID)
}

// TODO: RpcIDThreshold
func (s *stickyNodes) Success(d *StickyDetails) error {
	if d.StickyOptions.IsEmpty() {
		return nil
	}

	// TODO: check if Origin checking is required (it is already checked elsewhere)
	// We trust the passed StickyClient, to avoid additional cache look-ups
	sc, ok := s.items[d.Key]
	if !ok {
		newItem := &d.StickyClient
		newItem.SetupCounts()
	} else {
		d.StickyClient = sc
	}

	return s.enforceRelayLimit(d)
}

func (c *StickyClient) SetupCounts() {
	c.Relays = &CountWithTTL{}
	c.Errors = &CountWithTTL{}
}

// TODO: pocket-go may need to return the node that was actually used: alternatively, we could always set the preferred node when using pocket-go to send relays.
func (s *stickyNodes) enforceRelayLimit(d *StickyDetails) error {
	now := time.Now()

	sc := d.StickyClient
	sc.Relays.Count++

	if sc.Relays.Count > s.settings.RelayLimit {
		s.log.WithFields(logger.Fields{"StickyDetails": d}).Info("deleting entry due to relay limit")
		delete(s.items, d.Key)
		return nil
	}

	if sc.Relays.Count == 1 { // New Entry in Relay Count
		sc.Relays.TTL = now.Add(s.settings.Duration)
	}

	if sc.Relays.TTL.Before(now) {
		s.log.WithFields(logger.Fields{"StickyDetails": d}).Info("resetting relay counts since TTL has passed")
		sc.Relays.TTL = now.Add(s.settings.Duration)
		sc.Relays.Count = 0
	}

	s.items[d.Key] = sc
	return nil
}

func (s *stickyNodes) increaseErrorCount(d *StickyDetails) error {
	now := time.Now()

	sc := d.StickyClient
	sc.Errors.Count++

	if sc.Errors.Count > s.settings.MaxErrors {
		s.log.WithFields(logger.Fields{"StickyDetails": d}).Info("deleting entry due to error limit")
		delete(s.items, d.Key)
		return nil
	}

	if sc.Errors.Count == 1 { // New Entry in Errors Count
		sc.Errors.TTL = now.Add(s.settings.Duration)
	}

	if sc.Errors.TTL.Before(now) {
		s.log.WithFields(logger.Fields{"StickyDetails": d}).Info("resetting error counts since TTL has passed")
		sc.Errors.TTL = now.Add(s.settings.Duration)
		sc.Errors.Count = 0
	}

	s.items[d.Key] = sc
	return nil
}

func (c StickyClient) NodeMatches(nodeAddress string) bool {
	// TODO: getAddressFromPublicKey(node)
	return c.PreferredNodeAddress != "" && c.PreferredNodeAddress == nodeAddress
}

func (s *stickyNodes) Failure(d *StickyDetails) error {
	if d.StickyOptions.IsEmpty() {
		return nil
	}

	sc, ok := s.items[d.Key]
	if !ok {
		newItem := &d.StickyClient
		newItem.SetupCounts()
	} else {
		d.StickyClient = sc
	}

	return s.increaseErrorCount(d)
}

type CountWithTTL struct {
	Count int
	TTL   time.Time
}

// IsExpired verifies if the TTL of the counter has expired. If so, it resets the
//	counter, and returns true.
func (c *CountWithTTL) IsExpired(t time.Time) bool {
	if t.After(c.TTL) {
		c.Count = 0
		return true
	}
	return false
}

func (c *CountWithTTL) Increase(newTTL time.Time) {
	c.Count++
	c.TTL = newTTL
}
