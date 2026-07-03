# Gem Bridge

Gem Bridge is a local AI tooling bridge written in Go.

The goal of this project is to provide a secure local daemon that exposes controlled filesystem and development tools to an AI assistant running in a browser-based interface.

This project is designed around three core principles:

- **Local-first execution**: files and commands are handled on the user's machine.
- **Workspace isolation**: tools are restricted to an authorized project directory.
- **Tool-agnostic protocol**: the daemon should be usable by different AI frontends in the future.

## Current Features

- Safe workspace path resolution
- Directory listing
- File reading
- Protection against absolute paths and path traversal attacks
- Consistent JSON responses for tool execution results

## Project Structure

```text
.
├── cmd/
│   └── gem-bridge/
│       └── main.go
├── internal/
│   ├── security/
│   │   └── workspace.go
│   └── tools/
│       └── files.go
├── go.mod
└── README.md
```

## Usage

Run the project from the workspace root.

List files in the current workspace:

```bash
go run ./cmd/gem-bridge '{"tool":"listDirectory","path":"."}'
```

Read a file:

```bash
go run ./cmd/gem-bridge '{"tool":"readFile","path":"go.mod"}'
```

Attempting to access files outside the workspace is blocked:

```bash
go run ./cmd/gem-bridge '{"tool":"readFile","path":"../../.ssh/id_rsa"}'
```

Expected response:

```json
{
  "success": false,
  "error": "access outside the workspace is blocked"
}
```

When executed through `go run`, this failed request also returns `exit status 1`, which is expected behavior for the current CLI version.

## Security Model

Gem Bridge treats the workspace root as a security boundary.

All user-provided paths must be relative and are resolved through the security layer before any filesystem operation is executed. Absolute paths and path traversal attempts are blocked to prevent access to files outside the authorized workspace.

Examples of blocked paths:

```text
/etc/passwd
/home/user/.ssh/id_rsa
../../.env
../../.config
```

## Roadmap

- Add configurable workspace root
- Add file writing with explicit safety rules
- Add Git tools
- Add WebSocket transport for local development
- Add browser extension integration
- Add Native Messaging support
- Add approval flow for sensitive operations
- Add automated tests for security and tool behavior
- Add structured logging

## Why This Project Exists

Modern AI assistants are increasingly useful for software development, but browser-based AI interfaces usually do not have safe, direct access to a developer's local project files.

Gem Bridge explores a local-first approach where an AI assistant can request controlled actions through a small, auditable daemon running on the developer's own machine.

The long-term goal is to create a secure bridge between conversational AI tools and local development workflows without exposing the entire filesystem or relying on unsafe remote access.

## License

This project is currently under active development. A license will be added before the first public release.
