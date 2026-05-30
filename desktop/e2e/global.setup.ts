import { mkdirSync, writeFileSync, rmSync } from "node:fs";
import { join } from "node:path";

// Place test state OUTSIDE the desktop/ directory so Wails dev's file watcher
// doesn't detect changes and trigger rebuilds during E2E test runs.
export const E2E_DIR = join(import.meta.dirname, "..", "..", ".e2e-run");
const STATE_FILE = join(E2E_DIR, "state.json");

export interface E2EState {
  testDir: string;
  configPath: string;
  sourceDir: string;
  files: string[];
}

function generateTestFiles(sourceDir: string): string[] {
  const files: { name: string; content: string }[] = [
    {
      name: "Quarterly Business Report.pdf",
      content: "Q4 2025 revenue projections and analysis",
    },
    {
      name: "Meeting Notes - Jan 15.txt",
      content: "Discussed roadmap priorities for the next quarter",
    },
    {
      name: "Profile Photo.png",
      content: "fake png content placeholder",
    },
    {
      name: "TAX DOCUMENT 2025.pdf",
      content: "Annual tax filing document with sensitive information",
    },
    {
      name: "img_0042.jpg",
      content: "fake jpeg content placeholder",
    },
  ];

  for (const f of files) {
    writeFileSync(join(sourceDir, f.name), f.content, "utf-8");
  }

  return files.map((f) => f.name);
}

function writeTestConfig(configPath: string) {
  const config = {
    output: "",
    case: "snake",
    ai: {
      provider: "ollama",
      model: "mock-llama",
      api_key: "",
      max_tokens: 100,
      temperature: 0.2,
      vision: { enabled: false, max_image_size: "10MB" },
      prompt:
        "You are a desktop organizer that creates nice names for the files with their context. Please follow snake case naming convention. Only respond with the new name and the file extension. Do not change the file extension.",
    },
    file_handling: {
      max_size: "100MB",
      auto_approve: true,
      move_files: false,
      skip_dot_files: true,
    },
    content_extraction: {
      extract_text: true,
      extract_metadata: true,
      max_content_length: 2048,
      skip_large_files: false,
      read_context: true,
    },
    performance: {
      ai: { workers: 2, timeout: "60s", retries: 1 },
      file: { workers: 4, timeout: "30s", retries: 1 },
    },
    logging: { enabled: false, log_path: "" },
  };

  mkdirSync(join(configPath, ".."), { recursive: true });
  writeFileSync(configPath, JSON.stringify(config, null, 2), "utf-8");
}

export default function globalSetup() {
  // Clean up any previous run artifacts
  try {
    rmSync(E2E_DIR, { recursive: true, force: true });
  } catch {
    // ok if it doesn't exist
  }

  const sourceDir = join(E2E_DIR, "source");
  const configPath = join(E2E_DIR, "config", "config.json");

  mkdirSync(sourceDir, { recursive: true });

  const files = generateTestFiles(sourceDir);
  writeTestConfig(configPath);

  const state: E2EState = {
    testDir: E2E_DIR,
    configPath,
    sourceDir,
    files,
  };

  // Write state to file so tests can read it (globalSetup runs in a separate process)
  writeFileSync(STATE_FILE, JSON.stringify(state, null, 2), "utf-8");

  console.log(`[e2e:setup] test dir:   ${E2E_DIR}`);
  console.log(`[e2e:setup] config:     ${configPath}`);
  console.log(`[e2e:setup] source dir: ${sourceDir}`);
  console.log(`[e2e:setup] files:      ${files.length}`);
}
