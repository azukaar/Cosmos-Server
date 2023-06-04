package utils

import (
	"net"
	"errors"
	"strings"
	// "fmt"
	// "os"
)

func CheckDNS(url string) error {
	Log("CheckDNS: " + url)
	
	realHostname := GetMainConfig().HTTPConfig.Hostname
	realHostname = strings.Split(realHostname, ":")[0]
	
	
	ips, err := net.LookupIP(url)
	ipsReal, errReal := net.LookupIP(realHostname)

	if err != nil {
		return err
	}

	if errReal != nil {
		return errReal
	}

	ipCheck := ""
	ipReal := ""

	for _, ip := range ips {
		// if IPV4
		if ip.To4() != nil {
			ipCheck = ip.String()
			break
		}
	}

	for _, ip := range ipsReal {
		if ip.To4() != nil {
			ipReal = ip.String()
			break
		}
	}

	if ipCheck != ipReal {
		return errors.New("DNS mismatch, this endpoint does not seem to point to your server IP: " + ipCheck + " != " + ipReal)
	}

	return nil
}