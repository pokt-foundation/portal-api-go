package session

import (
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
)

type SessionManager interface {
	GetSession(Key) (*provider.Session, error)
}

func NewSessionManager(dispatchUrls []string) SessionManager {
	return &sessionManager{
		dispatchUrls: dispatchUrls,
	}
}

var defaultTTL = 60 * time.Second

// TODO: add options to session manager: retryAttempts, rejectSelfSignedCertificates, timeout

type cacheEntry struct {
	*provider.Session
	TTL time.Time
}

type sessionManager struct {
	dispatchUrls []string
	sessions     map[Key]cacheEntry
}

type Key struct {
	PublicKey    string
	BlockchainID string
}

func (s *sessionManager) GetSession(k Key) (*provider.Session, error) {
	cached, ok := s.sessions[k]
	if !ok || cached.TTL.After(time.Now()) {
		return s.newSession(k)
	}
	return cached.Session, nil
}

func (s *sessionManager) newSession(k Key) (*provider.Session, error) {
	rpcProvider := provider.NewProvider(s.dispatchUrls[0], s.dispatchUrls)
	rpcProvider.UpdateRequestConfig(0, time.Duration(20)*time.Second)
	r, err := rpcProvider.Dispatch(k.PublicKey, k.BlockchainID, nil)
	if err != nil {
		return &provider.Session{}, err
	}
	s.sessions[k] = cacheEntry{
		Session: r.Session,
		TTL:     time.Now().Add(defaultTTL),
	}
	return r.Session, nil
}
