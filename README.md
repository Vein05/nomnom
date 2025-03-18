# NomNom 

A powerful Go CLI tool for bulk renaming files using AI. NomNom helps you intelligently rename multiple files at once by analyzing their content and generating meaningful names using state-of-the-art AI models.

!["nomnom-pic"](nomnom.png)

## Features ✨

- 🗂️ **Bulk Processing**: Rename entire folders of files in one go
- 📄 **Smart Content Analysis**: Supports various file types including:
  - Text files
  - Documents (PDF, DOCX)
  - Presentations
  - Images (metadata)
  - Videos (metadata)
- 🤖 **AI-Powered**: Multiple AI provider options:
  - DeepSeek V3/R1 model
  - OpenRouter support (access to Claude, GPT-4, and more)
  - Local execution via Ollama
- 👀 **Preview Mode**: Review generated names before applying changes
- 🎯 **Flexible Naming**: Supports different casing options (snake_case, camelCase, kebab, pascal.)
- 🔒 **Safe Operations**: Creates a separate directory for renamed files
- 📝 **Operation Logging**: Detailed logs of all rename operations with:
  - Full absolute paths for reliable tracking
  - Original and new file locations
  - Operation timestamps and status
  - Support for reverting changes
- ↩️ **Revert Support**: Ability to undo rename operations using logged history

## Installation

### Prerequisites

1. Install Tesseract OCR:
   - macOS: `brew install tesseract`
   - Ubuntu/Debian: `sudo apt-get install tesseract-ocr`
   - Windows: Download installer from [UB Mannheim](https://github.com/UB-Mannheim/tesseract/wiki)

2. Install Go package:
```bash
go install github.com/yourusername/nomnom@latest
```

## Quick Start 🚀

1. Create a `config.json` file (or use the default one):
```json
{
  "output": "./renamed",
  "case": "snake", // camel, kebab, pascal
  "ai": {
    "provider": "deepseek",
    "model": "deepseek-v3",
    "api_key": "your-api-key"
  }
}
```

2. Run NomNom:
```bash
nomnom --dir /path/to/files
```

## Usage 📖

### Basic Command
```bash
nomnom --dir <directory> [flags]
```

### Available Flags

| Flag           | Short | Description                                     |
|---------------|--------|-------------------------------------------------|
| --dir         | -d    | Source directory containing files to rename      |
| --config      | -c    | Path to config file (default: "config.json")     |
| --auto-approve| -y    | Automatically approve changes                    |
| --dry-run     | -n    | Preview changes without renaming                 |
| --verbose     | -v    | Enable verbose logging                          |
| --log         | -l    | Enable operation logging (default: true)        |

### Configuration

NomNom uses a JSON configuration file with the following options:

- **Output Settings**
  - `output`: Output directory path
  - `case`: Naming case style (snake, camel, etc.)

- **AI Settings**
  - Provider selection (DeepSeek/Ollama/OpenRouter)
  - Model configuration
  - API keys

- **File Handling**
  - File type inclusion/exclusion
  - Size limits
  - Backup options

- **Content Extraction**
  - Text extraction settings
  - Metadata processing
  - Content length limits

- **Performance**
  - Worker count
  - Timeout settings
  - Retry configuration

See the full configuration example in `config.json`.

### AI Provider Examples

#### Using DeepSeek (Default)
```json
{
  "ai": {
    "provider": "deepseek",
    "model": "deepseek-v3",
    "api_key": "your-deepseek-api-key"
  }
}
```

#### Using OpenRouter
```json
{
  "ai": {
    "provider": "openrouter",
    "model": "anthropic/claude-3-opus-20240229",
    "api_key": "your-openrouter-api-key",
    "temperature": 0.7,
    "max_tokens": 100
  }
}
```

OpenRouter gives you access to various models including:
- `anthropic/claude-3-opus-20240229`
- `anthropic/claude-3-sonnet-20240229`
- `openai/gpt-4-turbo-preview`
- And many more! Check [OpenRouter's model list](https://openrouter.ai/docs#models) for all available options. I recommend using `google/gemini-2.0-flash-001`.

### Logging and Reverting

NomNom includes a robust logging system that tracks all file operations:

#### Logging Features
- Logs are stored in `.nomnom/logs` in your target directory
- Each session generates a unique log file (`changes_{timestamp}.json`)
- Logs contain detailed information about each operation:
  ```json
  {
    "session_id": "1234567890",
    "start_time": "2024-03-21T10:00:00Z",
    "end_time": "2024-03-21T10:01:00Z",
    "entries": [
      {
        "timestamp": "2024-03-21T10:00:30Z",
        "operation": "rename",
        "original_path": "/absolute/path/to/original/file.txt",
        "new_path": "/absolute/path/to/new/file.txt",
        "base_dir": "/absolute/path/to/original",
        "relative_path": "file.txt",
        "success": true
      }
    ]
  }
  ```

#### Revert Operations (Coming Soon)
The revert functionality will allow you to:
- List available operation logs
- View details of specific rename sessions
- Revert changes from a specific session
- Perform dry-run reverts to preview changes
- Selectively revert specific files

Stay tuned for the revert command implementation!

## Contributing ��

Contributions are welcome! Please feel free to submit a Pull Request.

## License 📄

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 
