package main

import (
	"encoding/json"
	"github.com/NickTVA/DnsMon/dnsresolver"
	"github.com/NickTVA/DnsMon/httphandlers"
	"github.com/newrelic/go-agent/v3/newrelic"
	"os"
	"strconv"
	"time"
)

var nrApp *newrelic.Application
var hostnames []string
var pollInterval int
var monitorName string

func main() {

	monitorName = os.Getenv("MONITOR_NAME")
	hostnames = getHostsFromEnv()
	pollInterval = setPollIntervalFromEnv()

	nrApp = initNewRelic()
	println("Waiting up to a minute for NR connection")
	nrApp.WaitForConnection(time.Minute)

	go httphandlers.SetupHealthCheckHTTP(nrApp)

	nrApp.RecordCustomEvent("DNSMonStarted", map[string]interface{}{
		"NumHosts":    len(hostnames),
		"MonitorName": monitorName,
	})

	go monitorDNS(hostnames)
	foreverBlockingTickMonitor()
}

func setPollIntervalFromEnv() int {
	pollInterval = 300
	pollIntervalpropstring := os.Getenv("POLL_INTERVAL")
	pollIntervalPropint, err := strconv.Atoi(pollIntervalpropstring)
	if err == nil && (len(pollIntervalpropstring) > 0) {
		if pollIntervalPropint >= 30 {
			println("Setting poll interval to " + strconv.Itoa(pollIntervalPropint) + " seconds.")
			pollInterval = pollIntervalPropint
		} else {
			println("Setting pollinterval to minimum of 30 seconds")
			pollInterval = 30
		}
	} else {
		println("Unable to read POLL_INTERVAL, setting to 5 minutes")
	}

	return pollInterval
}

func getHostsFromEnv() []string {
	hostnames := make([]string, 0)

	//Read hostnames from environment
	for i := 0; i < 100; i++ {

		hostname := os.Getenv("dns.host." + strconv.Itoa(i))
		if len(hostname) < 1 {
			continue
		}
		hostnames = append(hostnames, hostname)

	}

	if len(hostnames) < 1 {
		println("No hostnames in ENV")
		os.Exit(-1)
	}
	return hostnames
}

func monitorDNS(hostnames []string) {

	var lastrun time.Time
	for true {

		lastrun = time.Now()

		for _, hostname := range hostnames {
			dnsinfo := dnsresolver.GetDNSInfo(hostname)
			dnsinfo["MonitorName"] = monitorName
			nrApp.RecordCustomEvent("DnsMon", dnsinfo)
			bytes, _ := json.Marshal(dnsinfo)

			println(string(bytes))

		}

		durationSecs := time.Second * time.Duration(pollInterval)

		for true {
			if time.Now().After(lastrun.Add(durationSecs)) {
				break
			}

			time.Sleep(1 * time.Second)
		}

	}
}

func foreverBlockingTickMonitor() {

	for true {
		event := map[string]interface{}{
			"monitor_name": monitorName,
			"num_hosts":    len(hostnames),
		}

		println("Tick")

		nrApp.RecordCustomEvent("DNSMonTick", event)
		time.Sleep(3 * time.Minute)

	}
}

func initNewRelic() *newrelic.Application {
	newrelicKey := os.Getenv("NEWRELIC_KEY")
	if len(newrelicKey) < 1 {
		print("Must set NEWRELIC_KEY with NewRelic license key")
		os.Exit(-1)
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("DNSMon"),
		newrelic.ConfigLicense(newrelicKey),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if app == nil || err != nil {
		print("NewRelic Not Initialized")
		os.Exit(-1)
	}
	return app
}
