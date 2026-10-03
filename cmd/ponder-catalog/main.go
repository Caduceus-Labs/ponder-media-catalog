// Command ponder-catalog builds and publishes the public Ponder media catalog.
//
// It is a single binary exposing one command per pipeline job (ingest, publish,
// run-all). Individual jobs are started by an external scheduler; this process
// is not a long-running service.
package main

import (
	"os"

	"github.com/Caduceus-Labs/ponder-media-catalog/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
