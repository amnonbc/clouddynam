package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

type Config struct {
	User string
	Domain string
	ApiKey string
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
		fmt.Println(err)
		return err
	}
	// Fetch the zone ID for zone example.org
	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Fetch all DNS records for example.org
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		fmt.Println(err)
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
			log.Println(err)
			return err
		}
		log.Println("Set", domain, "to", ip)
	}
	return nil
}

func loadConfig() error {
	f, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&config)
}

func main() {
	err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(os.Stdout).Encode(config)
	ip, err := myIP()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("my ip is", ip)

	updateDomain(config.Domain, ip)
}
