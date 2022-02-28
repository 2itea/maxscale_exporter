## Overview

[![Go Report Card](https://goreportcard.com/badge/github.com/wirdgroup/maxscale_exporter)](https://goreportcard.com/report/github.com/wirdgroup/maxscale_exporter)

*As i am not a programmer, any help will be welcome to update/beautify the code.*

Maxscale v2.3.*

Working with Maxscale REST API.

Created to work with galera cluster.

Maxscale metrics:
- router infos (ie: readwritesplit)
  - State of a router
  - Total connections of a router
  - Current connections of a router
  - Total queries of a router
  - Total queries routed to the master
  - Total queries routed to the slave
  - Total queries routed
  - Number of read/write transactions
  - Number of read only transactions
  - Number of replayed transactions


- Node infos:
  - Total queries sent to a node
  - Total read queries sent to a node
  - Total write queries sent to a node
  - Average query session duration on a node
  - Average selects per session on a node
  - Current status of a node
  - Current master in galera

### Requirements:
Your MaxScale instance needs to have the REST API enabled so this can pull the stats from your MaxScale server. You can read [here](https://mariadb.com/kb/en/8088/) how to set this up.

## Installation
get [Go](https://golang.org/dl), set a `$GOPATH`, and run

    go get github.com/2itea/maxscale_exporter

## Use
Make sure `$GOPATH/bin` is in your `$PATH`.

    $ maxscale_exporter -h
    Usage of maxscale_exporter:
      -address string
        	address to get maxscale statistics from (default "admin:mariadb@127.0.0.1:8003")
      -port string
        	the port that the maxscale exporter listens on (default "9195")

### Example output

```
# HELP maxscale_exporter_total_scrapes Current total MaxScale scrapes
# TYPE maxscale_exporter_total_scrapes counter
maxscale_exporter_total_scrapes 1
# HELP maxscale_node_master Current master in galera
# TYPE maxscale_node_master gauge
maxscale_node_master{id="node1"} 0
maxscale_node_master{id="node2"} 0
maxscale_node_master{id="node3"} 1
# HELP maxscale_node_query_avg_sess_duration Average query session duration on a node
# TYPE maxscale_node_query_avg_sess_duration gauge
maxscale_node_query_avg_sess_duration{id="Read-Write-Service",node="node1"} 9.89293
maxscale_node_query_avg_sess_duration{id="Read-Write-Service",node="node2"} 9.89293
maxscale_node_query_avg_sess_duration{id="Read-Write-Service",node="node3"} 9.89293
# HELP maxscale_node_query_read Total read queries sent to a node
# TYPE maxscale_node_query_read gauge
maxscale_node_query_read{id="Read-Write-Service",node="node1"} 20737
maxscale_node_query_read{id="Read-Write-Service",node="node2"} 12959
maxscale_node_query_read{id="Read-Write-Service",node="node3"} 454
# HELP maxscale_node_query_selects_per_session Average selects per session on a node
# TYPE maxscale_node_query_selects_per_session gauge
maxscale_node_query_selects_per_session{id="Read-Write-Service",node="node1"} 44
maxscale_node_query_selects_per_session{id="Read-Write-Service",node="node2"} 27
maxscale_node_query_selects_per_session{id="Read-Write-Service",node="node3"} 20
# HELP maxscale_node_query_total Total queries sent to a node
# TYPE maxscale_node_query_total gauge
maxscale_node_query_total{id="Read-Write-Service",node="node1"} 20737
maxscale_node_query_total{id="Read-Write-Service",node="node2"} 12959
maxscale_node_query_total{id="Read-Write-Service",node="node3"} 9822
# HELP maxscale_node_query_write Total write queries sent to a node
# TYPE maxscale_node_query_write gauge
maxscale_node_query_write{id="Read-Write-Service",node="node1"} 0
maxscale_node_query_write{id="Read-Write-Service",node="node2"} 0
maxscale_node_query_write{id="Read-Write-Service",node="node3"} 9368
# HELP maxscale_node_status Current status of a node
# TYPE maxscale_node_status gauge
maxscale_node_status{id="node1"} 1
maxscale_node_status{id="node2"} 1
maxscale_node_status{id="node3"} 1
# HELP maxscale_router_connections Total connections of a router
# TYPE maxscale_router_connections gauge
maxscale_router_connections{id="Read-Write-Service"} 454
# HELP maxscale_router_connections_current The total number of processed events
# TYPE maxscale_router_connections_current gauge
maxscale_router_connections_current{id="Read-Write-Service"} 1
# HELP maxscale_router_queries Total queries of a router
# TYPE maxscale_router_queries gauge
maxscale_router_queries{id="Read-Write-Service"} 42156
# HELP maxscale_router_route_all Total queries routed
# TYPE maxscale_router_route_all gauge
maxscale_router_route_all{id="Read-Write-Service"} 454
# HELP maxscale_router_route_master Total queries routed to the master
# TYPE maxscale_router_route_master gauge
maxscale_router_route_master{id="Read-Write-Service"} 9368
# HELP maxscale_router_route_slave Total queries routed to the slave
# TYPE maxscale_router_route_slave gauge
maxscale_router_route_slave{id="Read-Write-Service"} 32788
# HELP maxscale_router_transactions_replayed Number of replayed transactions
# TYPE maxscale_router_transactions_replayed gauge
maxscale_router_transactions_replayed{id="Read-Write-Service"} 0
# HELP maxscale_router_transactions_ro Number of read only transactions
# TYPE maxscale_router_transactions_ro gauge
maxscale_router_transactions_ro{id="Read-Write-Service"} 0
# HELP maxscale_router_transactions_rw Number of read/write transactions
# TYPE maxscale_router_transactions_rw gauge
maxscale_router_transactions_rw{id="Read-Write-Service"} 0
# HELP maxscale_up Was the last scrape of MaxScale successful?
# TYPE maxscale_up gauge
maxscale_up 1
```
