package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	envPrefix   = "MAXSCALE_EXPORTER"
	metricsPath = "/metrics"
	namespace   = "maxscale"
)

// Flags for CLI invocation
var (
	address *string
	port    *string
)

type MaxScale struct {
	Address       string
	up            prometheus.Gauge
	totalScrapes  prometheus.Counter
	routerMetrics map[string]Metric
	nodeMetrics   map[string]Metric
}

type Metric struct {
	Desc      *prometheus.Desc
	ValueType prometheus.ValueType
}

var (
	routerLabelNames = []string{"id"}
	nodeLabelNames   = []string{"id", "node"}
	nodesLabelNames  = []string{"id"}
)

type metrics map[string]Metric

type routerList struct {
	Name string
}

type nodeList struct {
	Name string
}

var floatTempNode float64

var (
	routerMetrics = metrics{
		"connectionstt":         newDesc("router", "state", "State of a router", routerLabelNames, prometheus.GaugeValue),
		"connections":           newDesc("router", "connections", "Total connections of a router", routerLabelNames, prometheus.GaugeValue),
		"current_connections":   newDesc("router", "connections_current", "The total number of processed events", routerLabelNames, prometheus.GaugeValue),
		"queries":               newDesc("router", "queries", "Total queries of a router", routerLabelNames, prometheus.GaugeValue),
		"route_master":          newDesc("router", "route_master", "Total queries routed to the master", routerLabelNames, prometheus.GaugeValue),
		"route_slave":           newDesc("router", "route_slave", "Total queries routed to the slave", routerLabelNames, prometheus.GaugeValue),
		"route_all":             newDesc("router", "route_all", "Total queries routed", routerLabelNames, prometheus.GaugeValue),
		"rw_transactions":       newDesc("router", "transactions_rw", "Number of read/write transactions", routerLabelNames, prometheus.GaugeValue),
		"ro_transactions":       newDesc("router", "transactions_ro", "Number of read only transactions", routerLabelNames, prometheus.GaugeValue),
		"replayed_transactions": newDesc("router", "transactions_replayed", "Number of replayed transactions", routerLabelNames, prometheus.GaugeValue),
	}
	nodeMetrics = metrics{
		"total":                   newDesc("node", "query_total", "Total queries sent to a node", nodeLabelNames, prometheus.GaugeValue),
		"read":                    newDesc("node", "query_read", "Total read queries sent to a node", nodeLabelNames, prometheus.GaugeValue),
		"write":                   newDesc("node", "query_write", "Total write queries sent to a node", nodeLabelNames, prometheus.GaugeValue),
		"avg_sess_duration":       newDesc("node", "query_avg_sess_duration", "Average query session duration on a node", nodeLabelNames, prometheus.GaugeValue),
		"avg_selects_per_session": newDesc("node", "query_selects_per_session", "Average selects per session on a node", nodeLabelNames, prometheus.GaugeValue),
		"node_status":             newDesc("node", "status", "Current status of a node", nodesLabelNames, prometheus.GaugeValue),
		"node_master":             newDesc("node", "master", "Current master in galera", nodesLabelNames, prometheus.GaugeValue),
	}
)

func newDesc(subsystem string, name string, help string, variableLabels []string, t prometheus.ValueType) Metric {
	return Metric{
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, name),
			help, variableLabels, nil,
		), t}
}

func NewExporter(address string) (*MaxScale, error) {
	return &MaxScale{
		Address: address,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of MaxScale successful?",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_scrapes",
			Help:      "Current total MaxScale scrapes",
		}),
		routerMetrics: routerMetrics,
		nodeMetrics:   nodeMetrics,
	}, nil
}

// Describe describes all the metrics ever exported by the MaxScale exporter. It
// implements prometheus.Collector.
func (m *MaxScale) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range m.routerMetrics {
		ch <- m.Desc
	}

	for _, m := range m.nodeMetrics {
		ch <- m.Desc
	}

	ch <- m.up.Desc()
	ch <- m.totalScrapes.Desc()
}

// Collect fetches the stats from configured MaxScale location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (m *MaxScale) Collect(ch chan<- prometheus.Metric) {
	m.totalScrapes.Inc()

	var parseErrors = false

	if err := m.parseServices(ch); err != nil {
		parseErrors = true
		log.Print(err)
	}

	if err := m.parseServers(ch); err != nil {
		parseErrors = true
		log.Print(err)
	}

	if parseErrors {
		m.up.Set(0)
	} else {
		m.up.Set(1)
	}
	ch <- m.up
	ch <- m.totalScrapes
}

func (m *MaxScale) getStatistics(path string) []byte {

	url := "http://" + m.Address + "/v1" + path

	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Printf("Error while doing request %v: %v\n", path, err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Printf("Error while getting %v: %v\n", path, err)
	}

	return body

}

func serverUp(status string) (float64, float64) {
	if strings.Contains(status, "Down") {
		return 0, 0
	}
	if strings.Contains(status, "Running") {
		if strings.Contains(status, "Master") {
			return 1, 1
		} else {
			return 1, 0
		}
	}
	return 0, 0
}

func (m *MaxScale) parseServices(ch chan<- prometheus.Metric) error {

	routers := []routerList{
		routerList{"connections"},
		routerList{"current_connections"},
		routerList{"queries"},
		routerList{"route_master"},
		routerList{"route_slave"},
		routerList{"route_all"},
		routerList{"rw_transactions"},
		routerList{"ro_transactions"},
		routerList{"replayed_transactions"},
	}

	nodes := []nodeList{
		nodeList{"total"},
		nodeList{"read"},
		nodeList{"write"},
		nodeList{"avg_sess_duration"},
		nodeList{"avg_selects_per_session"},
	}

	body := m.getStatistics("/services")

	// Loop in different services
	jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		// Get id service
		idRouter, err := jsonparser.GetString(value, "id")

		if err != nil {
			fmt.Println(err)
		}

		jsonparser.ObjectEach(value, func(key []byte, value2 []byte, dataType jsonparser.ValueType, offset int) error {
			for _, router := range routers {
				if string(key) == string(router.Name) {
					floatTemp, err := strconv.ParseFloat(string(value2), 64)

					if err != nil {
						fmt.Println(err)
					}
					connectionsMetric := m.routerMetrics[router.Name]
					ch <- prometheus.MustNewConstMetric(
						connectionsMetric.Desc,
						connectionsMetric.ValueType,
						floatTemp,
						idRouter,
					)
				}

				if string(key) == "server_query_statistics" {

					jsonparser.ArrayEach(value2, func(value3 []byte, dataType jsonparser.ValueType, offset int, err error) {
						idNode, _ := jsonparser.GetString(value3, "id")
						jsonparser.ObjectEach(value3, func(key2 []byte, value4 []byte, dataType jsonparser.ValueType, offset int) error {
							for _, node := range nodes {
								if string(key2) == string(node.Name) {
									if string(node.Name) == "avg_sess_duration" {
										floatTempNode, err = strconv.ParseFloat(strings.TrimSuffix(string(value4), "s"), 64)
										if err != nil {
											fmt.Println(err)
										}

									} else {
										floatTempNode, err = strconv.ParseFloat(string(value4), 64)
										if err != nil {
											fmt.Println(err)
										}

									}

									if err != nil {
										fmt.Println(err)
									}

									nodeConnectionsMetric := m.nodeMetrics[node.Name]
									ch <- prometheus.MustNewConstMetric(
										nodeConnectionsMetric.Desc,
										nodeConnectionsMetric.ValueType,
										floatTempNode,
										idRouter, idNode,
									)

								}
							}
							return nil
						})

						value2 = nil
					})
				}

			}
			return nil
		}, "attributes", "router_diagnostics")

	}, "data")

	return nil
}

func (m *MaxScale) parseServers(ch chan<- prometheus.Metric) error {

	body := m.getStatistics("/servers")

	jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		// Get id service
		idRouter, err := jsonparser.GetString(value, "id")

		if err != nil {
			fmt.Println(err)
		}

		stringStateNode, err := jsonparser.GetString(value, "attributes", "state")

		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(idRouter)
		//fmt.Println(statusNode)

		statusNode, masterNode := serverUp(stringStateNode)

		nodeStatueConnectionsMetric := m.nodeMetrics["node_status"]
		ch <- prometheus.MustNewConstMetric(
			nodeStatueConnectionsMetric.Desc,
			nodeStatueConnectionsMetric.ValueType,
			statusNode,
			idRouter,
		)

		nodeMasterConnectionsMetric := m.nodeMetrics["node_master"]
		ch <- prometheus.MustNewConstMetric(
			nodeMasterConnectionsMetric.Desc,
			nodeMasterConnectionsMetric.ValueType,
			masterNode,
			idRouter,
		)

	}, "data")

	return nil
}

// strflag is like flag.String, with value overridden by an environment
// variable (when present). e.g. with address, the env var used as default
// is MAXSCALE_EXPORTER_ADDRESS, if present in env.
func strflag(name string, value string, usage string) *string {
	if v, ok := os.LookupEnv(envPrefix + strings.ToUpper(name)); ok {
		return flag.String(name, v, usage)
	}
	return flag.String(name, value, usage)
}

func main() {
	log.SetFlags(0)

	address = strflag("address", "admin:mariadb@127.0.0.1:8989", "address to get maxscale statistics from")
	port = strflag("port", "9195", "the port that the maxscale exporter listens on")
	flag.Parse()

	log.Print("Starting MaxScale exporter")
	log.Printf("Scraping MaxScale REST API at: %v", *address)
	exporter, err := NewExporter(*address)
	if err != nil {
		log.Fatalf("Failed to start maxscale exporter: %v\n", err)
	}

	prometheus.MustRegister(exporter)
	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>MaxScale Exporter</title></head>
			<body>
			<h1>MaxScale Exporter</h1>
			<p><a href="` + metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	log.Printf("Started MaxScale exporter, listening on port: %v", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
