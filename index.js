import express from "express";
import cors from "cors";

const app = express();
app.use(express.json());
app.use(cors());

const port = process.env.PORT || 3000;

// MCP discovery endpoint
app.get("/.well-known/ai-plugin.json", (req, res) => {
  res.json({
    schema_version: "v1",
    name_for_human: "WhatsApp MCP Bridge",
    name_for_model: "whatsapp_mcp",
    description_for_human: "Bridge between WhatsApp and Make using MCP.",
    description_for_model: "MCP server that enables sending and receiving WhatsApp messages.",
    auth: { type: "none" },
    api: {
      type: "openapi",
      url: `${req.protocol}://${req.get("host")}/openapi.json`,
    },
  });
});

// OpenAPI description (minimal)
app.get("/openapi.json", (req, res) => {
  res.json({
    openapi: "3.0.1",
    info: {
      title: "WhatsApp MCP Bridge",
      version: "1.0.0",
    },
    paths: {
      "/": {
        post: {
          summary: "MCP handshake",
          responses: { "200": { description: "OK" } },
        },
      },
    },
  });
});

// Base POST handler for MCP
app.post("/", (req, res) => {
  res.status(200).json({ ok: true, message: "MCP endpoint connected" });
});

app.listen(port, () => {
  console.log(`✅ MCP server is running on port ${port}`);
});
