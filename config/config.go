package config

import (
	"flag"
	"github.com/alrevuelta/eth-pools-metrics/pools"
	log "github.com/sirupsen/logrus"
	"os"
)

// By default the release is a custom build. CI takes care of upgrading it with
// go build -v -ldflags="-X 'github.com/alrevuelta/eth-pools-metrics/config.ReleaseVersion=x.y.z'"
var ReleaseVersion = "custom-build"

type Config struct {
	PoolName              string
	Network               string
	WithdrawalCredentials []string
	FromAddress           []string
	BeaconRpcEndpoint     string
	PrometheusPort        int
	Postgres              string
}

// custom implementation to allow providing the same flag multiple times
// --flag=value1 --flag=value2
type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func NewCliConfig() (*Config, error) {
	var fromAddress arrayFlags
	var withdrawalCredentials arrayFlags

	flag.Var(&withdrawalCredentials, "withdrawal-credentials", "Withdrawal credentials used in the deposits. Can be used multiple times")
	flag.Var(&fromAddress, "from-address", "Wallet addresses used to deposit. Can be used multiple times")

	var network = flag.String("network", "mainnet", "Ethereum 2.0 network mainnet|prater|pyrmont")
	var beaconRpcEndpoint = flag.String("beacon-rpc-endpoint", "localhost:4000", "Address:Port of a eth2 beacon node endpoint")
	var prometheusPort = flag.Int("prometheus-port", 9500, "Prometheus port to listen to")
	var version = flag.Bool("version", false, "Prints the release version and exits")
	var poolName = flag.String("pool-name", "required", "Name of the pool being monitored. If known, addresses are loaded by default (see known pools)")
	var postgres = flag.String("postgres", "", "Postgres db endpoit: postgresql://user:password@netloc:port/dbname (optional)")
	flag.Parse()

	if *version {
		log.Info("Version: ", ReleaseVersion)
		os.Exit(0)
	}

	if *poolName == "required" {
		log.Fatal("pool-name flag is required")
	}

	// If the pool name is known, override from-address
	preLoadedAddresses := pools.PoolsAddresses[*poolName]
	if len(preLoadedAddresses) != 0 {
		log.Info("The pool-name is known, overriding from-address")
		fromAddress = preLoadedAddresses
	} else {
		if len(fromAddress) == 0 && len(withdrawalCredentials) == 0 {
			log.Fatal("Either withdrawal-credentials or from-address must be populated")
		}
	}

	conf := &Config{
		PoolName:              *poolName,
		Network:               *network,
		BeaconRpcEndpoint:     *beaconRpcEndpoint,
		PrometheusPort:        *prometheusPort,
		WithdrawalCredentials: withdrawalCredentials,
		FromAddress:           fromAddress,
		Postgres:              *postgres,
	}
	logConfig(conf)
	return conf, nil
}

func logConfig(cfg *Config) {
	log.WithFields(log.Fields{
		"PoolName":              cfg.PoolName,
		"BeaconRpcEndpoint":     cfg.BeaconRpcEndpoint,
		"WithdrawalCredentials": cfg.WithdrawalCredentials,
		"FromAddress":           cfg.FromAddress,
		"Network":               cfg.Network,
		"PrometheusPort":        cfg.PrometheusPort,
		"Postgres":              cfg.Postgres,
	}).Info("Cli Config:")
}
