package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type Config struct {
	Domains []string
	ApiKey  string
}

const (
	ipv4info = "https://api.ipify.org?format=json"
	ipv6info = "https://api64.ipify.org?format=json"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// myIP returns the caller's current public IP.
func myIP(ctx context.Context, u string) (ip netip.Addr, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return ip, err
	}
	r, err := httpClient.Do(req)
	if err != nil {
		return ip, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return ip, fmt.Errorf("attempt to get my IP returned %s", r.Status)
	}
	var msg struct {
		Ip netip.Addr
	}
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		return ip, err
	}
	return msg.Ip, nil
}

// updateDomain updates the DNS A or AAAA record for domain to point to ip.
func updateDomain(apiKey, domain, recordType string, ip netip.Addr) error {
	api, err := cloudflare.NewWithAPIToken(apiKey)
	if err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return err
	}
	slog.Debug("resolved zone", "domain", domain, "zoneID", zoneID)

	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{Name: domain, Type: recordType})
	if err != nil {
		return err
	}
	slog.Debug("found DNS records", "domain", domain, "type", recordType, "count", len(records))
	if len(records) == 0 {
		slog.Warn("no DNS records found, nothing to update", "domain", domain, "type", recordType)
	}

	for _, r := range records {
		if r.Content == ip.String() {
			slog.Debug("record already up to date", "domain", domain, "type", recordType, "ip", ip)
			continue
		}
		slog.Info("updating DNS record", "domain", domain, "from", r.Content, "to", ip)
		r.Content = ip.String()
		if err := api.UpdateDNSRecord(zoneID, r.ID, r); err != nil {
			return err
		}
		slog.Info("updated DNS record", "domain", domain, "ip", ip)
	}
	return nil
}

func loadConfig(fn string) (Config, error) {
	f, err := os.Open(fn)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	var cfg Config
	return cfg, json.NewDecoder(f).Decode(&cfg)
}

func main() {
	cf := flag.String("cfg", "config.json", "config file")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	if *debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	slog.Debug("loading config", "file", *cf)
	cfg, err := loadConfig(*cf)
	if err != nil {
		slog.Error("loading config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()

	slog.Debug("fetching IPv4 address", "url", ipv4info)
	ipv4, err := myIP(ctx, ipv4info)
	if err != nil {
		slog.Error("getting IPv4 address", "err", err)
		os.Exit(1)
	}
	if !ipv4.Is4() {
		slog.Error("not an IPv4 address", "ip", ipv4)
		os.Exit(1)
	}
	slog.Debug("got IPv4 address", "ip", ipv4)

	slog.Debug("fetching IPv6 address", "url", ipv6info)
	ipv6, err := myIP(ctx, ipv6info)
	hasIPv6 := err == nil && ipv6.Is6()
	if err != nil {
		slog.Warn("could not get IPv6 address, skipping AAAA updates", "err", err)
	} else if !ipv6.Is6() {
		slog.Warn("no IPv6 connectivity detected, skipping AAAA updates", "ip", ipv6)
	} else {
		slog.Debug("got IPv6 address", "ip", ipv6)
	}

	if hasIPv6 {
		slog.Info("current IPs", "ipv4", ipv4, "ipv6", ipv6)
	} else {
		slog.Info("current IPs", "ipv4", ipv4)
	}

	for _, domain := range cfg.Domains {
		slog.Debug("updating domain", "domain", domain)
		if err := updateDomain(cfg.ApiKey, domain, "A", ipv4); err != nil {
			slog.Error("updating A record", "domain", domain, "err", err)
		}
		if hasIPv6 {
			if err := updateDomain(cfg.ApiKey, domain, "AAAA", ipv6); err != nil {
				slog.Error("updating AAAA record", "domain", domain, "err", err)
			}
		}
	}
}
