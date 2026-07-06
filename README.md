# Gem Bridge

[Leia em português brasileiro](./README.pt-br.md)

Gem Bridge is a local AI tooling bridge written in Go.

The goal of this project is to provide a secure local daemon that exposes controlled filesystem and development tools to an AI assistant running in a browser-based interface.

This project is designed around three core principles:

- **Local-first execution**: files and commands are handled on the user's machine.
- **Workspace isolation**: tools are restricted to an authorized project directory.
- **Tool-agnostic protocol**: the daemon should be usable by different AI frontends in the future.

## Current Features

- Configurable workspace root through `--workspace`
- Safe workspace path resolution
- Directory listing
- File reading
- Safe file creation through `writeFile`
- Protection against empty paths, absolute paths, path traversal, Windows drive paths, UNC paths, and symlink escapes
- Conservative write behavior that refuses to overwrite existing files
- File write size limit
- Consistent JSON responses for tool execution results
- Cross-platform CI on Linux, macOS, and Windows

## Documentation

- [Architecture](./docs/ARCHITECTURE.md)
- [Security Model](./docs/SECURITY_MODEL.md)
- [Project Context](./docs/PROJECT_CONTEXT.md)

Brazilian Portuguese versions are available for public-facing documentation:

- [README.pt-br.md](./README.pt-br.md)
- [ARCHITECTURE.pt-br.md](./docs/ARCHITECTURE.pt-br.md)
- [SECURITY_MODEL.pt-br.md](./docs/SECURITY_MODEL.pt-br.md)

## Project Structure

```text
.
├── .github/
│   └── workflows/
│       └── ci.yml
├── cmd/
│   └── gem-bridge/
│       └── main.go
├── docs/
│   ├── ARCHITECTURE.md
│   ├── ARCHITECTURE.pt-br.md
│   ├── PROJECT_CONTEXT.md
│   ├── SECURITY_MODEL.md
│   └── SECURITY_MODEL.pt-br.md
├── internal/
│   ├── security/
│   │   ├── workspace.go
│   │   └── workspace_test.go
│   └── tools/
│       ├── files.go
│       └── files_test.go
├── go.mod
├── README.md
└── README.pt-br.md
```

## Usage

Run the project from the workspace root.

List files in the current workspace:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"listDirectory","path":"."}'
```

Read a file:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"readFile","path":"go.mod"}'
```

Create a new file:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"writeFile","path":"notes.txt","content":"hello from gem bridge"}'
```

Read the created file:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"readFile","path":"notes.txt"}'
```

Attempting to overwrite an existing file is blocked in the current version:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"writeFile","path":"notes.txt","content":"overwrite"}'
```

Expected response:

```json
{
  "success": false,
  "error": "file already exists"
}
```

Attempting to access files outside the workspace is blocked:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"readFile","path":"../../.ssh/id_rsa"}'
```

Expected response:

```json
{
  "success": false,
  "error": "access outside the workspace is blocked"
}
```

When executed through `go run`, failed requests also return `exit status 1`, which is expected behavior for the current CLI version.

## Security Model

Gem Bridge treats the workspace root as a security boundary.

All user-provided paths must be relative and are resolved through the security layer before any filesystem operation is executed. Empty paths, absolute paths, path traversal attempts, Windows drive paths, UNC paths, and symlink escapes are blocked to prevent access outside the authorized workspace.

Examples of blocked paths:

```text
/etc/passwd
/home/user/.ssh/id_rsa
../../.env
../../.config
C:\Users\user\.ssh\id_rsa
C:/Users/user/.ssh/id_rsa
\\server\share\secret.txt
```

The current `writeFile` implementation is intentionally conservative. It creates new text files only, refuses to overwrite existing files, limits content size, and resolves paths through the shared workspace security layer before writing to disk.

## Continuous Integration

The repository includes a GitHub Actions workflow that runs on pushes and pull requests to `main`.

The CI validates:

```bash
go fmt ./...
go test ./...
```

on:

```text
ubuntu-latest
macos-latest
windows-latest
```

## Roadmap

- Add Git tools:
  - `gitStatus`
  - `gitDiff`
- Add structured request and response packages
- Add WebSocket transport for local development
- Add browser extension integration
- Add Native Messaging support
- Add approval flow for sensitive operations
- Add structured logging
- Expand safe file mutation rules when needed

## Why This Project Exists

Modern AI assistants are increasingly useful for software development, but browser-based AI interfaces usually do not have safe, direct access to a developer's local project files.

Gem Bridge explores a local-first approach where an AI assistant can request controlled actions through a small, auditable daemon running on the developer's own machine.

The long-term goal is to create a secure bridge between conversational AI tools and local development workflows without exposing the entire filesystem or relying on unsafe remote access.

## License

This project is currently under active development. A license will be added before the first public release.
