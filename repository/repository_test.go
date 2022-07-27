package repository

import (
	//"encoding/json"
	"fmt"
	//"io/ioutil"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	logger "github.com/sirupsen/logrus"
)

func TestGetBlockchain(t *testing.T) {
	// TODO: replace logger with testing.T (likely need to build a struct to mimic the interface)
	// TODO: copy (a subset of?) blockchains to testdata sub-directory
	repo, err := NewRepository("/tmp", logger.New())
	if err != nil {
		t.Fatalf("Error setting up the repository: %s", err)
	}

	testCases := []struct {
		name        string
		description string
		alias       string
		expected    Blockchain
		expectedErr error
	}{
		{
			name:        "filter blockchains",
			description: "Blockchains are filtered to return the one matching the blockchainID",
			alias:       "btc-mainnet",
			expected: Blockchain{
				ID:                "0002",
				Blockchain:        "btc-mainnet",
				BlockchainAliases: []string{"btc-mainnet"},
			},
		},
		{
			name:        "error returned for no matches",
			description: "passing an alias which matches no blockchains results in an error",
			alias:       "Foo",
			expectedErr: fmt.Errorf("No blockchains found matching foo"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetBlockchain(tc.alias)
			if err != nil {
				if tc.expectedErr == nil {
					t.Errorf("unexpected failure: %v", err)
				}
				if !reflect.DeepEqual(tc.expectedErr, err) {
					t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
				}
			}

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("unexpected value (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetApplication(t *testing.T) {
	repo, err := NewRepository("testdata", logger.New())
	if err != nil {
		t.Fatalf("Error setting up the repository: %v", err)
	}

	testCases := []struct {
		name        string
		description string
		id          string
		expected    Application
		expectedErr error
	}{
		{
			name:        "filter applications",
			description: "Applications are filtered to return the one matching the ID",
			id:          "4m14c3bttu6818e7kqz3181j",
			expected: Application{
				ID:          "4m14c3bttu6818e7kqz3181j",
				Name:        "TEST_DATA_Name_for_4m14c3bttu6818e7kqz3181j",
				Description: "TESTDATA_Description_4m14c3bttu6818e7kqz3181j",
				PublicPocketAccount: PublicPocketAccount{
					Address:   "7fce086ea7c04a16654916110d40d341899875ee",
					PublicKey: "7eaceace60765c8bb544038e14dc8c26455df82f55f8edd753a0459ff8361feb",
				},
				FreeTierApplicationAccount: FreeTierApplicationAccount{
					Address:   "7fce086ea7c04a16654916110d40d341899875ee",
					PublicKey: "7eaceace60765c8bb544038e14dc8c26455df82f55f8edd753a0459ff8361feb",
				},
				GatewaySettings: GatewaySettings{
					WhitelistOrigins:     []string{},
					WhitelistUserAgents:  []string{},
					WhitelistContracts:   []string{},
					WhitelistMethods:     []string{},
					WhitelistBlockchains: []string{},
				},
			},
		},
		{
			name:        "error returned for no matches",
			description: "passing an ID which matches no applications results in an error",
			id:          "foo",
			expectedErr: fmt.Errorf("No applications found matching foo"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetApplication(tc.id)
			if err != nil {
				if tc.expectedErr == nil {
					t.Errorf("unexpected failure: %v", err)
				}
				if !reflect.DeepEqual(tc.expectedErr, err) {
					t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
				}
			}

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("Unexpected value (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetLoadBalancer(t *testing.T) {
	repo, err := NewRepository("testdata", logger.New())
	if err != nil {
		t.Fatalf("Error setting up the repository: %v", err)
	}

	testCases := []struct {
		name        string
		description string
		id          string
		expected    LoadBalancer
		expectedErr error
	}{
		{
			name:        "filter loadbalancers",
			description: "Loadbalancers are filtered to return the one matching the ID",
			id:          "625d654e2f7a15003b5b4308",
			expected: LoadBalancer{
				ID:   "625d654e2f7a15003b5b4308",
				Name: "test",
			},
		},
		{
			name:        "error returned for no matches",
			description: "passing an ID which matches no loadbalancers results in an error",
			id:          "foo",
			expectedErr: fmt.Errorf("No loadbalancers found matching foo"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetLoadBalancer(tc.id)
			if err != nil {
				if tc.expectedErr == nil {
					t.Errorf("unexpected failure: %v", err)
				}
				if !reflect.DeepEqual(tc.expectedErr, err) {
					t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
				}
			}

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("Unexpected value (-want +got):\n%s", diff)
			}
		})
	}
}
