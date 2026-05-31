# Changelog

## [1.0.0] - 2026-05-30

### Desktop App

- **Wails desktop app** with full AI rename workflow (browse → scan → preview → rename)
- 8 live themes with instant preview switching
- Onboarding wizard for first-time setup
- Settings page: AI config (3 providers), file handling, content extraction, performance tuning, logging
- Real-time job progress with per-file status
- History view of past rename sessions
- Analytics dashboard with usage stats and activity chart
- Config source management (browse, create, switch config files)
- Font scale slider for UI text sizing

### CLI

- `nomnom rename` — AI-powered batch file renaming
- `nomnom setup` — interactive configuration wizard
- `nomnom analytics` — local usage analytics
- `nomnom revert` — roll back the most recent rename session
- `--move-files` flag for in-place rename (override copy-by-default)
- Vision toggle support for image-aware AI providers
- Rename plan caching in `.nomnom/cache/` for recovery
- Streaming AI responses with cancellation support

### Fixes

- Per-file extraction limits to prevent memory exhaustion on large files
- Session-based temp directories for preview isolation
- AI parameter normalization across DeepSeek, Ollama, and OpenRouter
- CI: consolidated workflows, bumped Go to 1.25, gated E2E on PRs
- macOS DMG packaging with Applications symlink
- Cross-platform E2E test suite with mock AI server

## [0.7.2] - 2026-04-16

- Content extraction streamlining and revert logging improvements
- Vision toggle support with AI parameter normalization
- Rename plan caching for recovery resilience
- Text extraction limits and session-based temp directories

## [0.7.1] - 2026-04-09

- Interactive setup command
- Analytics command for local usage stats
- App service and lifecycle cleanup
- README refresh with setup instructions and banner

## [0.7.0] - 2026-04-08

- CLI flow routed through presenter adapters
- Dedicated terminal presentation helpers
- Runtime reporter and approver interfaces
- Config defaults aligned with current behavior
- DeepSeek and Ollama dependency updates

## [0.6.0] and earlier

See GitHub releases for earlier versions.
