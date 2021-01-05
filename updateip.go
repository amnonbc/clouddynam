package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/cloudflare/cloudflare-go"
)

type Config struct {
	Domains []string
	ApiKey  string
}

var config Config

// myIP returns the callers current public IP.
func myIP(ipv6 bool) (net.IP, error) {
	u := "https://api.ipify.org?format=json"
	if ipv6 {
		u = "https://api64.ipify.org?format=json"
	}
	r, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("attempt to get my IP returned %s", r.Status)
	}
	defer r.Body.Close()
	var msg struct {
		Ip net.IP
	}
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		return nil, err
	}
	return msg.Ip, nil
}

// updateDomain updates the DNS A record for domain to point to ip.
func updateDomain(domain string, recordType string, ip net.IP) error {
	api, err := cloudflare.NewWithAPIToken(config.ApiKey)
	if err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return err
	}

	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{Name: domain, Type: recordType})
	if err != nil {
		return err
	}

	for _, r := range records {
		if r.Content == ip.String() {
			log.Println("Ip for", r.Name, "already is set as", ip)
			continue
		}
		r.Content = ip.String()
		err := api.UpdateDNSRecord(zoneID, r.ID, r)
		if err != nil {
			return err
		}
		log.Println("Set", domain, "to", ip)
	}
	return nil
}

func loadConfig(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&config)
}

func main() {
	cf := flag.String("cfg", "config.json", "config file")
	flag.Parse()
	err := loadConfig(*cf)
	if err != nil {
		log.Fatal(err)
	}

	ipv4, err := myIP(false)
	if err != nil {
		log.Fatal(err)
	}
	ipv6, err := myIP(true)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("my ip is", ipv4, ipv6)

	for _, domain := range config.Domains {
		err := updateDomain(domain, "A", ipv4)
		if err != nil {
			log.Println(err)
		}
		err = updateDomain(domain, "AAAA", ipv6)
		if err != nil {
			log.Println(err)
		}

	}
}
