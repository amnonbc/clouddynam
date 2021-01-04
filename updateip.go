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
func myIP() (net.IP, error) {
	r, err := http.Get("https://api.ipify.org?format=json")
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
func updateDomain(domain string, ip net.IP) error {
	api, err := cloudflare.NewWithAPIToken(config.ApiKey)
	if err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return err
	}

	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{Name: domain, Type: "A"})
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

	ip, err := myIP()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("my ip is", ip)

	for _, domain := range config.Domains {
		err := updateDomain(domain, ip)
		if err != nil {
			log.Println(err)
		}
	}
}
