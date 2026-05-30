import { createServer, type IncomingMessage, type ServerResponse } from "node:http";

const PORT = parseInt(process.env.MOCK_OLLAMA_PORT || "11435", 10);
const MODEL = process.env.MOCK_OLLAMA_MODEL || "mock-llama";

interface ChatMessage {
  role: string;
  content: string;
  images?: unknown[];
}

interface ChatRequest {
  model: string;
  messages: ChatMessage[];
  stream?: boolean;
  options?: Record<string, unknown>;
}

/**
 * Extract the original filename from the prompt context.
 * The Go backend sends file metadata in the user message.
 * We look for patterns like "Original: foo.txt" or paths.
 */
function extractOriginalName(messages: ChatMessage[]): string {
  const userMsg = messages.find((m) => m.role === "user");
  if (!userMsg) return "renamed_file";

  const content = userMsg.content;

  // Try "Original:" pattern
  const originalMatch = content.match(/Original(?: Name)?:\s*([^\n]+)/i);
  if (originalMatch) return originalMatch[1].trim();

  // Try "File:" pattern
  const fileMatch = content.match(/File:\s*([^\n]+)/i);
  if (fileMatch) return fileMatch[1].trim();

  // Try a path pattern like /path/to/file.txt
  const pathMatch = content.match(/([^/\s]+\.[a-zA-Z0-9]{2,6})(?:\s|$)/);
  if (pathMatch) return pathMatch[1].trim();

  return "renamed_file";
}

/**
 * Generate a deterministic snake_case rename from the original filename.
 */
function generateRename(original: string): string {
  const dotIndex = original.lastIndexOf(".");
  const ext = dotIndex >= 0 ? original.slice(dotIndex) : "";
  const base = dotIndex >= 0 ? original.slice(0, dotIndex) : original;

  // Convert to snake_case
  let slug = base
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_|_$/g, "")
    .replace(/__+/g, "_");

  if (!slug) slug = "renamed_file";

  return `${slug}${ext}`;
}

function parseBody(req: IncomingMessage): Promise<string> {
  return new Promise((resolve, reject) => {
    let data = "";
    req.on("data", (chunk: Buffer) => {
      data += chunk.toString();
    });
    req.on("end", () => resolve(data));
    req.on("error", reject);
  });
}

function sendJSON(res: ServerResponse, status: number, body: unknown) {
  res.writeHead(status, { "Content-Type": "application/json" });
  res.end(JSON.stringify(body));
}

let requestCount = 0;

const server = createServer(async (req, res) => {
  const url = req.url || "/";

  try {
    // Ollama tags endpoint — used by the desktop AI status check
    if (url === "/api/tags" && req.method === "GET") {
      sendJSON(res, 200, {
        models: [
          { name: `${MODEL}:latest`, modified_at: new Date().toISOString(), size: 1 },
        ],
      });
      return;
    }

    // Ollama chat endpoint — the actual rename API
    if (url === "/api/chat" && req.method === "POST") {
      const body = await parseBody(req);
      const chatReq: ChatRequest = JSON.parse(body);

      const original = extractOriginalName(chatReq.messages);
      const renamed = generateRename(original);
      requestCount++;

      console.log(
        `[mock-ollama] #${requestCount} "${original}" → "${renamed}" (model: ${chatReq.model})`,
      );

      sendJSON(res, 200, {
        model: chatReq.model || MODEL,
        created_at: new Date().toISOString(),
        message: {
          role: "assistant",
          content: renamed,
        },
        done: true,
        done_reason: "stop",
        total_duration: 50_000_000,
        load_duration: 1_000_000,
        prompt_eval_count: 10,
        prompt_eval_duration: 5_000_000,
        eval_count: 4,
        eval_duration: 20_000_000,
      });
      return;
    }

    // Health check
    if (url === "/health") {
      sendJSON(res, 200, { status: "ok", uptime: process.uptime() });
      return;
    }

    res.writeHead(404);
    res.end("not found");
  } catch (err) {
    console.error("[mock-ollama] error:", err);
    res.writeHead(500);
    res.end("internal error");
  }
});

server.listen(PORT, () => {
  console.log(`[mock-ollama] listening on http://localhost:${PORT}`);
  console.log(`[mock-ollama] model: ${MODEL}`);
});

// Graceful shutdown
process.on("SIGINT", () => {
  console.log("\n[mock-ollama] shutting down");
  server.close();
  process.exit(0);
});
process.on("SIGTERM", () => {
  server.close();
  process.exit(0);
});
