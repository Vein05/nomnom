# NomNom Desktop

This directory contains the Wails desktop app for NomNom.

## Layout

- `main.go`: desktop entry point
- `app.go`: Wails bridge methods and desktop state
- `frontend/`: React + TypeScript UI
- `build/`: generated desktop build output

The desktop app is a presentation layer. Core file, AI, logging, and analytics behavior stays in Go and is shared with the root CLI/backend packages.

## Development

From this directory:

```bash
wails dev
```

The frontend can also be developed separately:

```bash
cd frontend
npm install
npm run dev
```

## Build

```bash
wails build -clean
```

The production bundle is written under `build/bin/`.

CI also verifies native desktop production builds on macOS and Windows, and release builds publish those desktop artifacts alongside the CLI binaries.

## Notes

- The desktop backend uses the shared root service layer, so scans and runs follow the same backend path as the CLI.
- Desktop cancellation is wired through the shared apply pipeline and reports a real `canceled` job state.
- Keep the Wails bridge thin. If behavior affects files, jobs, persistence, or safety, it should live in Go rather than React.
