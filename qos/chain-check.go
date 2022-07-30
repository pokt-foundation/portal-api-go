package qos

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/pokt-foundation/pocket-go/relayer"

	logger "github.com/sirupsen/logrus"

	"github.com/pokt-foundation/portal-api-go/repository"
)

type ChainChecker interface {
	NodesSupportingApp(context.Context, *repository.Application, []*repository.Blockchain) (map[string][]*provider.Node, error)
}

type SessionRetriever func(context.Context, *repository.Application, string) (*provider.Session, error)

// TODO: pocket-go needs an interface so this can be removed
type PocketRelayer interface {
	Relay(*relayer.Input, *provider.RelayRequestOptions) (*relayer.Output, error)
}

func NewChainChecker(r PocketRelayer, s SessionRetriever, l *logger.Logger) (ChainChecker, error) {
	return nodeChecker{
		PocketRelayer:    r,
		SessionRetriever: s,
		Logger:           l,
	}, nil
}

type nodeChecker struct {
	PocketRelayer
	SessionRetriever
	*logger.Logger
}

// NodesSupportingApp verifies the nodes supporting each of the chains of an application, and returns the results
//	The results is a map of session keys to list of supporting nodes' addresses (public keys)
func (c nodeChecker) NodesSupportingApp(ctx context.Context, app *repository.Application, chains []*repository.Blockchain) (map[string][]*provider.Node, error) {
	pocketAAT := &provider.PocketAAT{
		AppPubKey:    app.GatewayAAT.ApplicationPublicKey,
		ClientPubKey: app.GatewayAAT.ClientPublicKey,
		Version:      app.GatewayAAT.Version,
		Signature:    app.GatewayAAT.ApplicationSignature,
	}

	var running int
	ch := make(chan *chainCheckResult, len(chains))
	for _, chain := range chains {
		running++
		go func(results chan *chainCheckResult, blockchain *repository.Blockchain) {
			log := c.Logger.WithFields(logger.Fields{"Application": app, "Chain": chain})
			session, err := c.SessionRetriever(ctx, app, chain.ChainID)
			if err != nil {
				log.WithFields(logger.Fields{"Error": err}).Warn("Error getting session")
				results <- nil
				return
			}

			nodes, err := c.nodesSupportingChain(pocketAAT, chain, session)
			if err != nil {
				log.WithFields(logger.Fields{"error": err}).Warn("Failed to check support for chain")
				results <- nil
				return
			}

			results <- &chainCheckResult{Nodes: nodes, Key: session.Key}
		}(ch, chain)
	}

	results := make(map[string][]*provider.Node)
	for running > 0 {
		r := <-ch
		running--
		if r == nil {
			continue
		}
		results[r.Key] = r.Nodes
	}

	return results, nil
}

type chainCheckResult struct {
	Nodes []*provider.Node
	Key   string
}

// TODO: pass a context to allow both Deadline and Cancellation
// supportingNodes returns the list of nodes in the session that supports the specified chain
//	returns the list of nodes' addresses (public keys)
func (c nodeChecker) nodesSupportingChain(aat *provider.PocketAAT, blockchain *repository.Blockchain, session *provider.Session) ([]*provider.Node, error) {
	// TODO: allow configuring max Parallelism at session level (if necessary)
	var count int
	ch := make(chan *provider.Node, len(session.Nodes))
	for _, n := range session.Nodes {
		count++
		go func(node *provider.Node, results chan<- *provider.Node) {
			supported, err := c.nodeSupportsChain(aat, blockchain, node, session)
			// TODO: log/report node's failure to process relay
			if err == nil && supported {
				results <- node
			} else {
				results <- nil
			}
		}(n, ch)
	}

	var supportingNodes []*provider.Node
	for count > 0 {
		n := <-ch
		if n != nil {
			supportingNodes = append(supportingNodes, n)
		}
		count--
	}
	return supportingNodes, nil
}

func (n nodeChecker) nodeSupportsChain(aat *provider.PocketAAT, blockchain *repository.Blockchain, node *provider.Node, session *provider.Session) (bool, error) {
	// TODO: Difference between blockchain.ChainID and blockchain.ID
	relay := relayer.Input{
		Method:     http.MethodPost,
		Blockchain: blockchain.ID,
		Data:       blockchain.ChainIDCheck,
		Path:       blockchain.Path,
		PocketAAT:  aat,
		Session:    session,
		Node:       node,
	}

	r, err := n.PocketRelayer.Relay(&relay, nil)
	if err != nil {
		return false, fmt.Errorf("Error relaying: %w", err)
	}

	chainID := r.RelayOutput.Response

	return chainID == blockchain.ChainID, nil
}

func pocketAAT(app repository.Application) provider.PocketAAT {
	return provider.PocketAAT{
		AppPubKey:    app.GatewayAAT.ApplicationPublicKey,
		ClientPubKey: app.GatewayAAT.ClientPublicKey,
		Version:      app.GatewayAAT.Version,
		Signature:    app.GatewayAAT.ApplicationSignature,
	}
}
