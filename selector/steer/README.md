# Steer selector

The Steer selector strategy should steer all requests for the given entropy to a single node for the given key, or if that node is failing, a consistent second-choice etc.

It tries to consistently steer all requests for a given set of entropy to a single instance to improve caching memory efficiency.

It should steer all requests for the given entropy, regardless of source service, to a single node, or if that node is failing, a consistent second-choice, etc.

This differs from a filter based selector in that the input set of nodes is not restricted, but rather the entire available set of nodes are processed in a preferential ordering.

# Re-balancing requests

When a new node appears, it will get a fair share of requests randomly allocated from across the existing services as the new node will then be higher scoring for approximately `1/count(nodes)` of the ids.

Similarly, when a node disappears, its load will get fairly redistributed amongst the existing remaining nodes.

# Benefits

This benefits us in that memory can be more optimally used by trying to target requests that have support for caching to servers that are more likely to have that data already, whilst not requiring all servers to have all data for everything cached.

Over time, this results in overall memory savings where one is running multiple services, whist still allowing for fractional re-balancing with auto-scaling and unexpected node failure/replacement, without resorting to a more rigid sharding type strategy.

# Usage

This method is a call option, which can be passed into client RPC requests.

## Example

```go
rsp, err := myClient.ClientCall(
    ctx,
    &ClientCallRequest{
    	//...
        SomeID: id,
    },
    steer.Strategy(id),
)
```
