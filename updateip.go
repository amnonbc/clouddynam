package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

type Config struct {
	User    string
	Domains []string
	ApiKey  string
}

var config Config

type ipMsg struct {
	Ip net.IP
}

func myIP() (net.IP, error) {
	r, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return nil, err
	}
	if r.StatusCode != 200 {
		return nil, fmt.Errorf("attempt to get my IP returned %s", r.Status)
	}
	defer r.Body.Close()
	var msg ipMsg
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		return nil, err
	}
	return msg.Ip, nil
}

func updateDomain(domain string, ip net.IP) error {
	api, err := cloudflare.New(config.ApiKey, config.User)
	if err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return err
	}

	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		return err
	}

	for _, r := range records {
		if r.Type != "A" {
			continue
		}
		if r.Name != domain {
			continue
		}
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
