package dnsresolver

import (
	"net"
	"strconv"
	"time"
)

func GetDNSInfo(host string) map[string]interface{} {

	dnsinfo := map[string]interface{}{
		"dns_host": host,
	}

	cname, err := net.LookupCNAME(host)

	dnsinfo["cname_error"] = 0
	dnsinfo["cname"] = ""
	if err != nil {
		dnsinfo["cname_error"] = 1
		dnsinfo["cname_error_message"] = err.Error()

	} else {
		dnsinfo["cname"] = cname
	}

	dnsinfo["dns_error"] = 0
	start := time.Now()
	ips, err := net.LookupIP(host)
	duration := time.Since(start)
	dnsinfo["duration_milliseconds"] = duration.Milliseconds()
	if err != nil {
		dnsinfo["dns_error"] = 1
		dnsinfo["error_message"] = err.Error()
	}
	for i, ip := range ips {

		dnsinfo["ip_"+strconv.Itoa(i)] = ip.String()
	}

	return dnsinfo
}
