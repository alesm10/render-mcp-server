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
		mcpHandler, err := server.NewStreamableHTTPServer(
			s,
			server.WithHTTPContextFunc(multicontext.MultiHTTPContextFunc(
				session.ContextWithHTTPSession(sessionStore),
				authn.ContextWithAPITokenFromHeader,
			)),
		)
		if err != nil {
			log.Fatalf("Failed to initialize MCP HTTP server: %v", err)
		}
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
	}

	

