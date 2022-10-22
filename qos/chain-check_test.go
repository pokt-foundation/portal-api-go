//revive:disable
package qos

// TODO: uncomment when tests pass

// func TestNodeSupportsChain(t *testing.T) {
// 	testCases := []struct {
// 		name        string
// 		blockchain  repository.Blockchain
// 		node        *provider.Node
// 		response    *relayer.Output
// 		relayError  error
// 		expected    bool
// 		expectedErr error
// 	}{
// 		{
// 			name: "Returns true when node supports the chain",
// 			response: &relayer.Output{
// 				RelayOutput: &provider.RelayOutput{
// 					Response: "0x12",
// 				},
// 			},
// 			blockchain: repository.Blockchain{
// 				ChainID: "18", // 0x12
// 			},
// 			expected: true,
// 		},
// 		{
// 			name: "Returns false when node does not support the chain",
// 			response: &relayer.Output{
// 				RelayOutput: &provider.RelayOutput{
// 					Response: "0x12",
// 				},
// 			},
// 			blockchain: repository.Blockchain{
// 				ChainID: "1000", // != 0x12
// 			},
// 		},
// 		{
// 			name:        "Relay error results in error",
// 			relayError:  fmt.Errorf("Error sending relay"),
// 			expectedErr: fmt.Errorf("Error sending relay"),
// 		},
// 		{
// 			name: "Invalid relay response results in error",
// 			response: &relayer.Output{
// 				RelayOutput: &provider.RelayOutput{
// 					Response: "foo",
// 				},
// 			},
// 			expectedErr: fmt.Errorf("Error parsing"),
// 		},
// 	}

// 	aat := &provider.PocketAAT{
// 		AppPubKey:    "applicationPublicKey",
// 		ClientPubKey: "clientPublicKey",
// 		Version:      "version",
// 		Signature:    "applicationSignature",
// 	}
// 	session := &provider.Session{
// 		Nodes: []*provider.Node{
// 			{Address: "node-1"},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			fakeRelayer := &fakePocketRelayer{
// 				responses: map[string]*relayer.Output{"node-1": tc.response},
// 				errors:    map[string]error{"node-1": tc.relayError},
// 			}
// 			nodeChecker := nodeChecker{
// 				PocketRelayer: fakeRelayer,
// 			}

// 			got, err := nodeChecker.nodeSupportsChain(aat, &tc.blockchain, &provider.Node{Address: "node-1"}, session)

// 			if tc.expectedErr != nil {
// 				// TODO: use errors.Is (needs custom errors defined)
// 				if err == nil || !strings.Contains(err.Error(), tc.expectedErr.Error()) {
// 					t.Fatalf("Expected error: %v, got: %v", tc.expectedErr, err)
// 				}
// 				return
// 			}

// 			if got != tc.expected {
// 				t.Errorf("Expected %t, got %t", tc.expected, got)
// 			}
// 			if diff := cmp.Diff(aat, fakeRelayer.relay.PocketAAT); diff != "" {
// 				t.Errorf("unexpected Pocket AAT (-want +got):\n%s", diff)
// 			}
// 			if diff := cmp.Diff(session, fakeRelayer.relay.Session); diff != "" {
// 				t.Errorf("unexpected Session (-want +got):\n%s", diff)
// 			}
// 		})
// 	}
// }

// func TestNodesSupportingChain(t *testing.T) {
// 	blockchain := repository.Blockchain{ChainID: "18"} // 0x12
// 	aat := &provider.PocketAAT{
// 		AppPubKey:    "applicationPublicKey",
// 		ClientPubKey: "clientPublicKey",
// 		Version:      "version",
// 		Signature:    "applicationSignature",
// 	}
// 	session := &provider.Session{
// 		Nodes: []*provider.Node{
// 			{Address: "node-1"},
// 			{Address: "node-2"},
// 			{Address: "node-3"},
// 		},
// 	}

// 	testCases := []struct {
// 		name      string
// 		responses map[string]*relayer.Output
// 		errors    map[string]error
// 		expected  []*provider.Node
// 	}{
// 		{
// 			name: "All nodes of the session are checked",
// 			responses: map[string]*relayer.Output{
// 				"node-1": {RelayOutput: &provider.RelayOutput{Response: "0x12"}},
// 				"node-2": {RelayOutput: &provider.RelayOutput{Response: "0x12"}},
// 				"node-3": {RelayOutput: &provider.RelayOutput{Response: "0x12"}},
// 			},
// 			expected: []*provider.Node{
// 				{Address: "node-1"},
// 				{Address: "node-2"},
// 				{Address: "node-3"},
// 			},
// 		},
// 		{
// 			name: "Node not supporting a chain is not returned",
// 			responses: map[string]*relayer.Output{
// 				"node-1": {RelayOutput: &provider.RelayOutput{Response: "0x12"}},
// 				"node-2": {RelayOutput: &provider.RelayOutput{Response: "0x1"}},
// 				"node-3": {RelayOutput: &provider.RelayOutput{Response: "0x12"}},
// 			},
// 			expected: []*provider.Node{
// 				{Address: "node-1"},
// 				{Address: "node-3"},
// 			},
// 		},
// 		{
// 			name: "Node with failed relay is not returned",
// 			responses: map[string]*relayer.Output{
// 				"node-1": {RelayOutput: &provider.RelayOutput{Response: "0x12"}},
// 				"node-3": {RelayOutput: &provider.RelayOutput{Response: "0x12"}},
// 			},
// 			errors: map[string]error{
// 				"node-2": fmt.Errorf("Error relaying"),
// 			},
// 			expected: []*provider.Node{
// 				{Address: "node-1"},
// 				{Address: "node-3"},
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			fakeRelayer := &fakePocketRelayer{
// 				responses: tc.responses,
// 				errors:    tc.errors,
// 			}
// 			nodeChecker := nodeChecker{
// 				PocketRelayer: fakeRelayer,
// 			}

// 			got, _ := nodeChecker.nodesSupportingChain(aat, &blockchain, session)
// 			sort.Slice(got, func(i, j int) bool {
// 				return got[i].Address < got[j].Address
// 			})

// 			if diff := cmp.Diff(tc.expected, got); diff != "" {
// 				t.Errorf("unexpected nodes (-want +got):\n%s", diff)
// 			}
// 		})
// 	}
// }

// type fakePocketRelayer struct {
// 	relay     *relayer.Input
// 	responses map[string]*relayer.Output
// 	errors    map[string]error
// }

// func (f *fakePocketRelayer) Relay(relay *relayer.Input, options *provider.RelayRequestOptions) (*relayer.Output, error) {
// 	f.relay = relay
// 	return f.responses[relay.Node.Address], f.errors[relay.Node.Address]
// }
