# DNS Selector

The dns selector looks up dns SRV records

- Service Node Id and Port are set as the Target
- The default domain is `micro.local`

## Usage

```go
selector := dns.NewSelector()

service := micro.NewService(
	micro.Selector(selector),
)
```

Specify lookup domain

```go
dns.NewSelector(
	dns.Domain("example.com"),
)
```
