package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/render-oss/render-mcp-server/cmd"
	"github.com/render-oss/render-mcp-server/pkg/cfg"
)

func main() {
	// Define and parse command line flags
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	flag.BoolVar(versionFlag, "v", false, "Print version information and exit")

	var transport string
	flag.StringVarP(&transport, "transport", "t", "", "Transport type (stdio or http)")
	flag.Parse()

	// ✅ If not provided via flag, read from ENV
	if transport == "" {
		if envTransport := os.Getenv("TRANSPORT"); envTransport != "" {
			transport = envTransport
		} else {
			transport = "stdio"
		}
	}

	if *versionFlag {
		fmt.Println("render-mcp-server version", cfg.Version)
		os.Exit(0)
	}

	fmt.Printf("🚀 Starting Render MCP Server with transport: %s\n", transport)
	cmd.Serve(transport)
}
