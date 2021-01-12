# Clouddynam

Update cloudflare DNS.

![Go](https://github.com/amnonbc/clouddynam/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/amnonbc/clouddynam)](https://goreportcard.com/report/github.com/amnonbc/clouddynam)
[![Go Reference](https://pkg.go.dev/badge/github.com/amnonbc/clouddynam.svg)](https://pkg.go.dev/github.com/amnonbc/clouddynam)

This utility may be useful if you are using Cloudflare as a dydnamic DNS service. The utility will 
check the current external IP address and then update Cloudflare's DNS to 
direct traffic to your external IP.

The utility requires a config file:

```
{
    "Domains": ["yourdomain.com"],
    "ApiKey": "ApiKeyFromCloudFlare"
}
```

You can run the utility as a cron.

## Building

`go build`
