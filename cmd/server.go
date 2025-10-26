package cmd

import (
	"log"
	"net/http"
	"os"
	"encoding/json"

	"github.com/mark3labs/mcp-go/server"
	"github.com/render-oss/render-mcp-server/pkg/authn"
	"github.com/render-oss/render-mcp-server/pkg/cfg"
	"github.com/render-oss/render-mcp-server/pkg/client"
	"github.com/render-oss/render-mcp-server/pkg/deploy"
	"github.com/render-oss/render-mcp-server/pkg/keyvalue"
	"github.com/render-oss/render-mcp-server/pkg/logs"
	"github.com/render-oss/render-mcp-server/pkg/metrics"
	"github.com/render-oss/render-mcp-server/pkg/multicontext"
	"github.com/render-oss/render-mcp-server/pkg/owner"
	"github.com/render-oss/render-mcp-server/pkg/postgres"
	"github.com/render-oss/render-mcp-server/pkg/service"
	"github.com/render-oss/render-mcp-server/pkg/session"
)

func Serve(transport string) *server.MCPServer {
	// Create MCP server
	s := server.NewMCPServer(
		"render-mcp-server",
		cfg.Version,
	)

	c, err := client.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	s.AddTools(owner.Tools(c)...)
	s.AddTools(service.Tools(c)...)
	s.AddTools(deploy.Tools(c)...)
	s.AddTools(postgres.Tools(c)...)
	s.AddTools(keyvalue.Tools(c)...)
	s.AddTools(logs.Tools(c)...)
	s.AddTools(metrics.Tools(c)...)

	// --- ✅ JSON-RPC health endpoint ---
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]interface{}{
				"ok":      true,
				"message": "render-mcp-server running",
			},
		})
	})
	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	// --- ✅ konec health endpointu ---

	if transport == "http" {
		var sessionStore session.Store
		if redisURL, ok := os.LookupEnv("REDIS_URL"); ok {
			log.Print("using Redis session store\n")
			sessionStore, err = session.NewRedisStore(redisURL)
			if err != nil {
				log.Fatalf("failed to initialize Redis session store: %v", err)
			}
		} else {
			log.Print("using in-memory session store\n")
			sessionStore = session.NewInMemoryStore()
		}
		err := server.
			NewStreamableHTTPServer(s, server.WithHTTPContextFunc(multicontext.MultiHTTPContextFunc(
				session.ContextWithHTTPSession(sessionStore),
				authn.ContextWithAPITokenFromHeader,
			))).
			Start(":10000")
		if err != nil {
			log.Fatalf("Starting Streamable server: %v\n:", err)
		}
	} else {
		err := server.ServeStdio(s, server.WithStdioContextFunc(multicontext.MultiStdioContextFunc(
			session.ContextWithStdioSession,
			authn.ContextWithAPITokenFromConfig,
		)))
		if err != nil {
			log.Fatalf("Starting STDIO server: %v\n", err)
		}
	}

	return s
}

