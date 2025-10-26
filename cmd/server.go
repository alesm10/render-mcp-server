package cmd

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

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
	s := server.NewMCPServer("render-mcp-server", cfg.Version)

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

		mux := http.NewServeMux()

		// ✅ Health check endpoint
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

		// ✅ MCP API endpoint
		mcpHandler := server.NewStreamableHTTPHandler(s, server.WithHTTPContextFunc(
			multicontext.MultiHTTPContextFunc(
				session.ContextWithHTTPSession(sessionStore),
				authn.ContextWithAPITokenFromHeader,
			),
		))
		mux.Handle("/mcp", mcpHandler)

		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}

		log.Printf("✅ MCP server listening on port %s\n", port)
		err = http.ListenAndServe(":"+port, mux)
		if err != nil {
			log.Fatalf("Server error: %v\n", err)
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


