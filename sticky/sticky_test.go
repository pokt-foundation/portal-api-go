package sticky

import (
	"testing"
	"time"

	logger "github.com/sirupsen/logrus"

	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	maxRelays = 5
	maxErrors = 5
	duration  = 60 * time.Second
)

var (
	key = Key{
		ApplicationID: "app-1",
		BlockchainID:  "block-1",
		IP:            "123.456.789.001",
	}
	now = time.Now()
)

func TestEnforceRelayLimit(t *testing.T) {
	testCases := []struct {
		name     string
		existing map[Key]StickyClient
		input    *StickyDetails
		duration time.Duration
		expected *CountWithTTL
	}{
		{
			name:     "New entry should have maximum TTL on relay limits",
			existing: make(map[Key]StickyClient),
			input: &StickyDetails{
				Key:          key,
				StickyClient: relayData(0, time.Time{}),
			},
			duration: duration,
			expected: &CountWithTTL{
				Count: 1,
				TTL:   time.Now().Add(duration),
			},
		},
		{
			name: "Existing entry with TTL left should be increased",
			existing: map[Key]StickyClient{
				key: relayData(3, now.Add(10*duration)),
			},
			input: &StickyDetails{
				Key:          key,
				StickyClient: relayData(3, now.Add(10*duration)),
			},
			duration: duration,
			expected: &CountWithTTL{
				Count: 4,
				TTL:   now.Add(10 * duration),
			},
		},
		{
			name: "Expired entry should be removed",
			existing: map[Key]StickyClient{
				key: relayData(2, now.Add(-1*duration)),
			},
			input: &StickyDetails{
				Key:          key,
				StickyClient: relayData(2, now.Add(-1*duration)),
			},
			duration: duration,
			expected: &CountWithTTL{
				Count: 0,
				TTL:   now.Add(duration),
			},
		},
		{
			name: "Entry is removed if relays go above maximum allowed",
			existing: map[Key]StickyClient{
				key: relayData(5, now.Add(10*duration)),
			},
			input: &StickyDetails{
				Key:          key,
				StickyClient: relayData(5, now.Add(10*duration)),
			},
			duration: duration,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sn := stickyNodes{
				items: tc.existing,
				settings: StickyNodeSettings{
					Duration:   tc.duration,
					RelayLimit: maxRelays,
				},
				log: logger.New(),
			}
			_ = sn.enforceRelayLimit(tc.input)

			var all []StickyClient
			for _, sc := range sn.items {
				all = append(all, sc)
			}

			if tc.expected == nil {
				if len(all) > 0 {
					t.Fatalf("Expected no sticky items, found: %v", all)
				}
				return
			}
			if len(all) != 1 {
				t.Fatalf("Expected exactly 1 sticky item, found: %v", all)
			}

			actual := all[0].Relays
			if actual == nil {
				t.Fatalf("Expected Relay Count %v, got nil", tc.expected)
			}
			if actual.Count != tc.expected.Count {
				t.Errorf("Expected Relay Count: %d, got: %d", tc.expected.Count, actual.Count)
			}

			if actual.TTL.Before(tc.expected.TTL) {
				t.Errorf("Expected TTL after %s, got: %s", tc.expected.TTL, actual.TTL)
			}
		})
	}
}

func relayData(count int, ttl time.Time) StickyClient {
	return StickyClient{
		PreferredApplicationID: "app-1-ID",
		PreferredNodeAddress:   "node-1",
		Errors: &CountWithTTL{
			Count: count,
			TTL:   ttl,
		},
		Relays: &CountWithTTL{
			Count: count,
			TTL:   ttl,
		},
	}
}

func TestFailure(t *testing.T) {
	testCases := []struct {
		name     string
		existing map[Key]StickyClient
		duration time.Duration
		input    *StickyDetails
		expected *CountWithTTL
	}{
		{
			name:     "New entry with error count of 1 is added if no existing ones found",
			existing: make(map[Key]StickyClient),
			duration: duration,
			input: &StickyDetails{
				Key:           key,
				StickyOptions: repository.StickyOptions{Stickiness: true, StickyOrigins: []string{"origin-1"}},
				StickyClient:  relayData(0, time.Time{}),
			},
			expected: &CountWithTTL{Count: 1, TTL: now.Add(duration)},
		},
		{
			name:     "No entry saved if Stickiness option is not set",
			existing: make(map[Key]StickyClient),
			duration: duration,
			input: &StickyDetails{
				Key:           key,
				StickyOptions: repository.StickyOptions{},
				StickyClient:  relayData(0, now),
			},
		},
		{
			name: "Existing entry gets error count updated",
			existing: map[Key]StickyClient{
				key: relayData(2, now.Add(duration)),
			},
			duration: duration,
			input: &StickyDetails{
				Key:           key,
				StickyOptions: repository.StickyOptions{Stickiness: true, StickyOrigins: []string{"origin-1"}},
				StickyClient:  relayData(2, now.Add(duration)),
			},
			expected: &CountWithTTL{Count: 3, TTL: now.Add(duration)},
		},
		{
			name: "Error limit is enforced by removing the violating element",
			existing: map[Key]StickyClient{
				key: relayData(5, now.Add(10*duration)),
			},
			input: &StickyDetails{
				Key:           key,
				StickyOptions: repository.StickyOptions{Stickiness: true, StickyOrigins: []string{"origin-1"}},
				StickyClient:  relayData(5, now.Add(10*duration)),
			},
			duration: duration,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sn := stickyNodes{
				items: tc.existing,
				settings: StickyNodeSettings{
					Duration:   tc.duration,
					RelayLimit: maxRelays,
					MaxErrors:  maxErrors,
				},
				log: logger.New(),
			}

			_ = sn.Failure(tc.input)

			var all []StickyClient
			for _, sc := range sn.items {
				all = append(all, sc)
			}
			if tc.expected == nil {
				if len(all) > 0 {
					t.Fatalf("Expected 0 cached items, found: %v", all)
				}
				return
			}
			if len(all) != 1 {
				t.Fatalf("Expected 1 cached item, found: %d: %v", len(all), all)
			}
			actual := all[0].Errors

			if tc.expected.Count != actual.Count {
				t.Errorf("Expected Error count to be %d, got %d", tc.expected.Count, actual.Count)
			}

			if actual.TTL.Before(tc.expected.TTL) {
				t.Errorf("Expected TTL for relay errors to be after %v, got: %v", tc.expected.TTL, actual.TTL)
			}
		})
	}

}

func TestSuccess(t *testing.T) {
	testCases := []struct {
		name     string
		existing map[Key]StickyClient
		duration time.Duration
		input    *StickyDetails
		expected StickyClient
	}{
		{
			name:     "New entry added if no existing ones found",
			existing: make(map[Key]StickyClient),
			duration: duration,
			input: &StickyDetails{
				Key:           key,
				StickyOptions: repository.StickyOptions{Stickiness: true, StickyOrigins: []string{"origin-1"}},
				StickyClient:  relayData(0, time.Time{}),
			},
			expected: relayData(1, time.Time{}),
		},
		{
			name:     "No entry saved if Stickiness option is not set",
			existing: make(map[Key]StickyClient),
			duration: duration,
			input: &StickyDetails{
				Key:           key,
				StickyOptions: repository.StickyOptions{},
				StickyClient:  relayData(0, time.Time{}),
			},
		},
		{
			name: "Existing entry gets relay count updated",
			existing: map[Key]StickyClient{
				key: relayData(2, now.Add(duration)),
			},
			duration: duration,
			input: &StickyDetails{
				Key:           key,
				StickyOptions: repository.StickyOptions{Stickiness: true, StickyOrigins: []string{"origin-1"}},
				StickyClient:  relayData(2, now.Add(duration)),
			},
			expected: relayData(3, time.Time{}),
		},
		{
			name: "Error limit is enforced by removing the violating element",
			existing: map[Key]StickyClient{
				key: relayData(maxRelays, now.Add(duration)),
			},
			duration: duration,
			input: &StickyDetails{
				Key:           key,
				StickyOptions: repository.StickyOptions{Stickiness: true, StickyOrigins: []string{"origin-1"}},
				StickyClient:  relayData(maxRelays, now.Add(duration)),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sn := stickyNodes{
				items: tc.existing,
				settings: StickyNodeSettings{
					Duration:   tc.duration,
					RelayLimit: maxRelays,
				},
				log: logger.New(),
			}

			_ = sn.Success(tc.input)

			var all []StickyClient
			for _, sc := range sn.items {
				all = append(all, sc)
			}
			if tc.expected.IsEmpty() {
				if len(all) > 0 {
					t.Fatalf("Expected 0 cached items, found: %v", all)
				}
				return
			}
			if len(all) != 1 {
				t.Fatalf("Expected 1 cached item, found: %d: %v", len(all), all)
			}
			actual := all[0]

			if tc.expected.Relays.Count != actual.Relays.Count {
				t.Errorf("Expected Relay count to be %d, got %d", tc.expected.Relays.Count, actual.Relays.Count)
			}
		})
	}
}
