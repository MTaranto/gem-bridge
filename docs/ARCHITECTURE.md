# Architecture

[Leia em português brasileiro](./ARCHITECTURE.pt-br.md)

## Purpose

Gem Bridge is a local-first tooling bridge written in Go.

Its purpose is to expose controlled local development tools to browser-based AI assistants without giving unrestricted access to the user's machine. The daemon is designed to run locally, enforce a strict workspace boundary, and remain independent from any single AI provider or browser frontend.

## Architectural Goals

Gem Bridge follows these architectural goals:

1. **Local-first execution**
   - Files, Git operations, and development commands run on the user's own machine.
   - The daemon should not require remote execution or unsafe cloud access to operate.

2. **Workspace isolation**
   - The configured workspace root is the main security boundary.
   - User-provided paths must be relative.
   - Absolute paths, path traversal, and symlink escapes must be blocked before any filesystem operation runs.

3. **Tool-agnostic protocol**
   - The daemon should not depend on a specific AI provider.
   - Browser assistants, browser extensions, local scripts, WebSocket clients, or Native Messaging hosts should be able to reuse the same tool model.

4. **Cross-platform readiness**
   - The project should be able to support Linux, macOS, and Windows.
   - Platform-specific behavior must be isolated when it becomes necessary.
   - The core tool and security model should remain shared across platforms.

5. **Incremental evolution**
   - The architecture should stay simple while the project is small.
   - Abstractions should be introduced when there is a real need, not before.
   - Security-sensitive behavior must be documented and tested as the project evolves.

## Current Architecture

The current version is a CLI-based prototype.

```text
JSON request
    ↓
cmd/gem-bridge
    ↓
tool dispatcher
    ↓
internal/tools
    ↓
internal/security
    ↓
authorized workspace
```

The CLI receives a JSON request, creates a restricted workspace, dispatches the requested tool, and returns a consistent JSON response.

Current supported tools:

- `listDirectory`
- `readFile`

Current request example:

```json
{
  "tool": "readFile",
  "path": "go.mod"
}
```

Current success response example:

```json
{
  "success": true,
  "data": "..."
}
```

Current error response example:

```json
{
  "success": false,
  "error": "access outside the workspace is blocked"
}
```

## Package Boundaries

The project should evolve around clear package responsibilities.

```text
cmd/gem-bridge
    CLI entry point, argument parsing, request decoding, response encoding.

internal/security
    Workspace boundary enforcement, path validation, symlink safety, future command safety helpers.

internal/tools
    User-facing tools exposed to AI clients, such as file reads, directory listing, future file writing, and future Git operations.

internal/config
    Future local configuration loading and persistence.

internal/transport
    Future transports such as WebSocket, HTTP-local, and Native Messaging.

internal/platform
    Future OS-specific behavior that cannot remain portable through the Go standard library.
```

The project should prefer domain-based boundaries first. Platform-specific packages should only be introduced when a feature truly requires different behavior on Linux, macOS, or Windows.

## Workspace Security Model

The workspace root is the primary security boundary.

All user-controlled paths must go through the security layer before any filesystem operation touches the disk.

The daemon must reject:

- Empty paths
- Absolute paths
- Path traversal attempts
- Paths escaping the workspace through symlinks
- Unsafe future write targets
- Unsafe future command execution requests

Examples of blocked paths:

```text
/etc/passwd
/home/user/.ssh/id_rsa
../../.env
../../.config
C:\Users\user\.ssh\id_rsa
\\server\share\secret.txt
```

The security layer should remain reusable by all transports. A request coming from CLI, WebSocket, Native Messaging, or any future frontend must receive the same path validation.

## Cross-Platform Strategy

Gem Bridge should support Linux, macOS, and Windows, but platform differences should be handled carefully and incrementally.

### Filesystem paths

Path handling is security-sensitive and must be tested across platforms.

Linux and macOS absolute paths usually look like:

```text
/home/user/project
/etc/passwd
```

Windows absolute paths may look like:

```text
C:\Users\User\project
C:/Users/User/project
\\server\share
```

The code should prefer Go's `filepath` package for OS-aware path behavior. Cross-platform tests should be added through CI so path safety is validated on Linux, macOS, and Windows.

### Local configuration

Future configuration should use the operating system's standard config directory through `os.UserConfigDir()`.

Expected examples:

```text
Linux:  ~/.config/gem-bridge/config.json
macOS:  ~/Library/Application Support/gem-bridge/config.json
Windows: %APPDATA%\gem-bridge\config.json
```

### Native Messaging

Future browser extension support may use Native Messaging.

Native Messaging host registration is platform-specific, so it should be isolated in a dedicated package when implemented.

Possible future package:

```text
internal/platform/nativehost
```

### Command execution

Gem Bridge should not expose unrestricted shell execution.

Instead of accepting arbitrary commands, the daemon should expose explicit tools such as:

- `gitStatus`
- `gitDiff`
- `goTest`
- `goFmt`

Each tool should call known binaries with controlled arguments. This keeps behavior safer and easier to support across Linux, macOS, and Windows.

## Transport Strategy

The current transport is CLI-based.

Future transports should reuse the same request, response, tool, and security layers.

Expected transport evolution:

```text
CLI prototype
    ↓
WebSocket transport for local development
    ↓
Browser extension integration
    ↓
Native Messaging support
```

A failed tool request should not crash a future long-running daemon. In CLI mode, returning exit status `1` for failed requests is acceptable. In daemon mode, errors should be returned as structured JSON while the process keeps running.

## Tool Design Principles

Tools exposed to AI clients must be small, explicit, and auditable.

A good tool:

- Has a clear name
- Has a narrow purpose
- Receives structured input
- Validates input before acting
- Uses the security layer for filesystem access
- Returns structured JSON
- Avoids hidden side effects

A bad tool:

- Accepts arbitrary shell commands
- Accepts absolute paths from the client
- Performs broad filesystem access
- Mutates files without clear rules
- Mixes transport, business logic, and security checks in the same place

## Future File Writing Rules

File writing will be more dangerous than file reading and must have explicit safety rules.

Future write operations should:

- Require relative paths
- Resolve through the workspace security layer
- Block path traversal and symlink escapes
- Refuse writes outside the workspace
- Avoid overwriting files unless explicitly allowed
- Consider size limits
- Return clear JSON errors
- Eventually support approval flow for sensitive operations

## Future Git Tool Rules

Git tools should be explicit and controlled.

Safe early Git tools:

- `gitStatus`
- `gitDiff`

More sensitive future Git tools:

- `gitAdd`
- `gitCommit`
- `gitCheckout`
- `gitMerge`
- `gitPush`

Sensitive Git operations should require additional validation and possibly user approval.

## Testing Strategy

Security-sensitive behavior must be covered by automated tests.

Current priority areas:

- Workspace root creation
- Relative path resolution
- Empty path rejection
- Absolute path rejection
- Path traversal rejection
- Symlink escape rejection

Future priority areas:

- Cross-platform path behavior on Linux, macOS, and Windows
- Safe file writing rules
- Git tool behavior
- Transport-level request handling
- Native Messaging registration helpers
- Command allowlist behavior

The project should eventually run CI on:

```text
ubuntu-latest
macos-latest
windows-latest
```

## Non-Goals for the Current Stage

The current stage should avoid:

- Arbitrary shell execution
- Full GUI automation
- Remote daemon exposure
- Complex plugin systems
- Premature OS adapter abstractions
- Unrestricted filesystem access
- Sensitive operations without explicit rules

## Design Decision: Avoid Premature OS Abstraction

Gem Bridge should be cross-platform-ready, but it should not start with a large generic OS abstraction layer.

Avoid introducing broad interfaces such as:

```go
type OSAdapter interface {
    ResolvePath(path string) (string, error)
    RunCommand(name string, args ...string) ([]byte, error)
    ConfigDir() (string, error)
    RegisterNativeHost() error
}
```

This kind of abstraction would be too broad for the current stage and would likely hide security decisions behind vague platform methods.

Instead, the project should introduce focused abstractions only when real platform-specific needs appear.

## Long-Term Vision

The long-term architecture is:

```text
Browser AI assistant
    ↓
Browser extension
    ↓
Gem Bridge local daemon
    ↓
controlled tools
    ↓
authorized workspace
```

Gem Bridge should become a small, auditable, local-first bridge that allows AI assistants to interact with development workspaces safely, without exposing the entire filesystem and without requiring users to migrate fully into a specific IDE or AI tool.
