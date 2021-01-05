# Clouddynam

Update cloudflare DNS.

[![Go Report Card](https://goreportcard.com/badge/github.com/amnonbc/clouddynam)](https://goreportcard.com/report/github.com/amnonbc/clouddynam)

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
