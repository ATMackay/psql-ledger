package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/ATMackay/psql-ledger/cmd/config"
	"github.com/ATMackay/psql-ledger/service"
)

const envPrefix = "PSQLLEDGER"

var (
	configFilePath string
	configFilePtr  = flag.String("config", "config.yml", "path to config file")
)

// RUN WITH PLAINTEXT CONFIG [RECOMMENDED FOR TESTING ONLY]
// $ go run main.go --config ./config.yml
// $ go run main.go --config {path_to_config_file}
//
// OR RUN WITH ENVIRONMENT VARIABLES
//
// $ go build
// $ export PSQLLEDGER_PASSWORD=<your_password>
// $ ./fusionkms
//
//

func init() {
	// Parse flag containing path to config file
	flag.Parse()
	if configFilePtr != nil {
		configFilePath = *configFilePtr
	}
}

func main() {
	var cfg service.Config

	if err := config.ParseYAMLConfig(configFilePath, &cfg, envPrefix); err != nil {
		panic(fmt.Sprintf("error parsing config: %v", err))
	}

	psqlLedger, err := service.BuildService(cfg)
	if err != nil {
		panic(fmt.Sprintf("error building service: %v", err))
	}

	psqlLedger.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	psqlLedger.Stop(sig)
}
