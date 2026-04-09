# NomNom

NomNom is a Go CLI for renaming files with AI. It scans a directory, extracts lightweight context from each file, asks an AI model for a better filename, and writes the renamed copies into a separate output directory.

![NomNom logo](nomnom.png)

## What It Does

- Renames files from a selected directory without modifying the originals
- Supports preview mode before applying changes
- Organizes output into category folders when enabled
- Logs rename sessions to `.nomnom/logs`
- Reverts a previous session into `.nomnom/reverted/<session_id>`

## Current File Support

- Text and data: `txt`, `md`, `json`
- Documents via `go-fitz` text extraction: `pdf`, `docx`, `epub`, `pptx`, `xlsx`, `xls`
- Images: `png`, `jpg`, `jpeg`, `webp`
- Media metadata: `mp3`, `ogg`, `mp4`, `flac`, `m4a`, `dsf`, `wav`

Image renaming works best with a multimodal model. Document extraction currently uses text extraction from the first two pages, not OCR.

## Requirements

- Go 1.24+
- One configured AI provider:
  - DeepSeek
  - OpenRouter
  - Ollama

## Install

### Build from source

```bash
git clone https://github.com/vein05/nomnom.git
cd nomnom
go build .
```

### Config path

- macOS/Linux: `~/.config/nomnom/config.json`
- Windows: `%APPDATA%\nomnom\config.json`

Copy the example config first:

```bash
mkdir -p ~/.config/nomnom
cp config.example.json ~/.config/nomnom/config.json
```

## Config Notes

- `ai.provider` must be one of `deepseek`, `openrouter`, or `ollama`
- `ai.model` must be set explicitly for OpenRouter and Ollama
- If `ai.api_key` is empty:
  - DeepSeek will use `DEEPSEEK_API_KEY`
  - OpenRouter will use `OPENROUTER_API_KEY`
- `output` defaults to `<input>/nomnom/renamed`
- Logs are written under `.nomnom/logs` in the selected input directory

## Quick Start

Preview:

```bash
nomnom -d /path/to/files
```

Apply:

```bash
nomnom -d /path/to/files --dry-run=false
```

Use a built-in prompt:

```bash
nomnom -d /path/to/files -p research
nomnom -d /path/to/files -p images
```

Use a custom prompt directly:

```bash
nomnom -d /path/to/files -p "Rename papers by topic and venue in snake case."
```

Revert a session:

```bash
nomnom --revert /path/to/.nomnom/logs/changes_<timestamp>.json
```

## Flags

| Flag | Short | Description | Default |
| --- | --- | --- | --- |
| `--dir` | `-d` | Source directory to process | required unless using `--revert` |
| `--config` | `-c` | Config file path | OS default config path |
| `--auto-approve` | `-y` | Skip approval prompts | `false` |
| `--dry-run` | `-n` | Preview only | `true` |
| `--log` | `-l` | Write session logs | `true` |
| `--organize` | `-o` | Organize files by category | `true` |
| `--prompt` | `-p` | Built-in prompt name or custom prompt text | empty |
| `--revert` | `-r` | Revert from a log file | empty |

## Example Config

```json
{
  "output": "",
  "case": "snake",
  "ai": {
    "api_key": "",
    "provider": "openrouter",
    "model": "google/gemini-2.0-flash-001",
    "vision": {
      "enabled": true,
      "max_image_size": "10MB"
    },
    "max_tokens": 1000,
    "temperature": 0.7,
    "prompt": "You are a helpful assistant that renames files based on file content. Return only the new filename with the original extension in snake case."
  },
  "file_handling": {
    "max_size": "100MB",
    "auto_approve": false
  },
  "performance": {
    "ai": {
      "workers": 5,
      "timeout": "30s",
      "retries": 3
    },
    "file": {
      "workers": 5,
      "timeout": "30s",
      "retries": 1
    }
  },
  "logging": {
    "enabled": true,
    "log_path": ".nomnom/logs"
  }
}
```

## Testing

Run the full test suite:

```bash
go test ./...
```

Some provider integration tests are skipped unless their required environment variables are present.

## Status

This repository is currently a CLI-first codebase. The core packages are now separated by responsibility:

- `internal/ai`
- `internal/content`
- `internal/files`
- `internal/utils`

That layout is intended to make a later Wails frontend easier to build without dragging terminal-specific behavior into the app layer.
