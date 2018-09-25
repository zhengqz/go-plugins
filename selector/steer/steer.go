package steer

import (
	"strings"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
	"github.com/minio/highwayhash"
)

// zeroKey is the base key for all hashes, it is 32 zeros.
var zeroKey [32]byte

// Strategy returns a call option which tries to consistently steer all requests for a given set of entropy to a
// single instance to improve memory efficiency where instances are caching data.
//
// This is the preferred usage as it gives the ultimate flexibility for determining the entropy used.
//
// Usage:
//    `myClient.MyCall(ctx, req, steer.Strategy(req.ID))`
func Strategy(entropy ...string) client.CallOption {
	return client.WithSelectOption(NewSelector(entropy))
}

// NewSelector returns a `SelectOption` that directs all request according to the given `entropy`.
func NewSelector(entropy []string) selector.SelectOption {
	return selector.WithStrategy(func(services []*registry.Service) selector.Next {
		return Next(entropy, services)
	})
}

// Next returns a `Next` function which returns the next highest scoring node.
func Next(entropy []string, services []*registry.Service) selector.Next {
	possibleNodes, scores := ScoreNodes(entropy, services)

	return func() (*registry.Node, error) {
		var best uint64
		pos := -1

		// Find the best scoring node from those available.
		for i, score := range scores {
			if score >= best && possibleNodes[i] != nil {
				best = score
				pos = i
			}
		}

		if pos < 0 {
			// There was no node found.
			return nil, nil
		}

		// Choose this node and set it's score to zero to stop it being selected again.
		node := possibleNodes[pos]
		possibleNodes[pos] = nil
		scores[pos] = 0
		return node, nil
	}
}

// ScoreNodes returns a score for each node found in the given services.
func ScoreNodes(entropy []string, services []*registry.Service) (possibleNodes []*registry.Node, scores []uint64) {
	// Generate a base hashing key based off the supplied entropy values.
	key := highwayhash.Sum([]byte(strings.Join(entropy, ":")), zeroKey[:])

	// Get all the possible nodes for the services, and assign a hash-based score to each of them.
	for _, s := range services {
		for _, n := range s.Nodes {
			// Use the base key from above to calculate a derivative 64 bit hash number based off the instance ID.
			score := highwayhash.Sum64([]byte(n.Id), key[:])
			scores = append(scores, score)
			possibleNodes = append(possibleNodes, n)
		}
	}
	return
}
