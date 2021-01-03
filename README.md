# Clouddynam

Update cloudflare DNS.

This utility may be useful if you are using Cloudflare to proxy a website which
you are hosting on a home network with a dynamic IP. The utility will 
check the current external IP address and then update Cloudflare's DNS to 
direct traffic to your external IP.

The utility requires a config file:

```
{
"User": "your@email.com",
"Domains": ["yourdomain.com"],
"ApiKey": "ApiKeyFromCloudFlare"
}
```

You can run the utility as a cron.