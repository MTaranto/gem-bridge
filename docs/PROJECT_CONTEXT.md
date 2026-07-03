# Project Context

## Project Name

Gem Bridge

## Purpose

Gem Bridge is a local-first AI tooling bridge written in Go.

The goal is to build a secure local daemon that acts as a bridge between browser-based AI assistants and the user's local development environment. The daemon should expose controlled tools for filesystem access, development workflows, Git operations, and future browser integration without giving unrestricted access to the machine.

## Current Repository

- GitHub repository: `MTaranto/gem-bridge`
- Main module: `github.com/mtaranto/gem-bridge`
- Main branch: `main`
- Current working branch: `main`

## Current Status

The project currently has a CLI-based Go prototype with:

- `cmd/gem-bridge/main.go`
- `internal/security/workspace.go`
- `internal/security/workspace_test.go`
- `internal/tools/files.go`
- `docs/PROJECT_CONTEXT.md`
- `docs/ARCHITECTURE.md`
- `README.md`
- `.gitignore`
- `go.mod`

The initial commit was created and pushed to GitHub:

```text
feat: initialize secure local bridge
```

Pull Request #1 was opened from `feat/configurable-workspace` into `main`, reviewed, and merged.

The merged feature branch included:

```text
feat: add configurable workspace root
docs: add project context
```

After the PR merge, the architecture overview was added directly to `main`:

```text
docs: add architecture overview
```

The latest confirmed local state:

- `main` is aligned with `origin/main`.
- Working tree is clean.
- Latest commit on `main`: `docs: add architecture overview`.

## Core Principles

Gem Bridge should follow these principles:

1. **Local-first execution**
   - Files and commands should be handled on the user's machine.
   - The system should avoid unnecessary remote execution or unsafe cloud dependencies.

2. **Workspace isolation**
   - The daemon must only operate inside an authorized workspace root.
   - Absolute paths must be rejected.
   - Path traversal must be blocked.
   - Symlink escapes must be detected and blocked.

3. **Tool-agnostic protocol**
   - The daemon should not be tightly coupled to one AI provider.
   - The long-term goal is to support browser-based assistants through a neutral local protocol.

4. **Security-first design**
   - File reads, writes, Git operations, and command execution must be explicitly controlled.
   - Dangerous shell commands must be blocked.
   - Sensitive operations should eventually require user approval.

5. **Portfolio-quality engineering**
   - The project should evolve incrementally.
   - Git history should stay clean.
   - Feature branches should be used for code, security-sensitive changes, and larger documentation updates.
   - Commits should follow Conventional Commits with lowercase descriptions.

## Development Guidelines

Use professional Go practices:

- Keep code idiomatic and simple.
- Prefer clear package boundaries.
- Write comments in English for exported types, exported functions, and important security-related logic.
- Prefer small, focused functions.
- Avoid overengineering early.
- Add tests for security-sensitive behavior.
- Run formatting and tests before committing.

Before each commit or merge, validate changes with:

```bash
go fmt ./...
go test ./...
git diff
```

Use Conventional Commits, for example:

```text
feat: add configurable workspace root
fix: block symlink workspace escape
docs: update project context
docs: add architecture overview
test: add workspace security tests
```

## Current Architecture

The current architecture is a simple CLI-based prototype:

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

Current supported tools:

- `listDirectory`
- `readFile`

The current request format is:

```json
{
  "tool": "readFile",
  "path": "go.mod"
}
```

The current response format is:

```json
{
  "success": true,
  "data": "..."
}
```

Errors are returned as:

```json
{
  "success": false,
  "error": "access outside the workspace is blocked"
}
```

In the current CLI version, failed requests may also return exit status `1`. This is acceptable for CLI behavior. In a future daemon/server mode, failed tool requests should not crash the daemon.

## Current Workspace Behavior

Implemented behavior:

- Added `--workspace` CLI flag.
- Default workspace root is `.`.
- Workspace root is resolved to an absolute path.
- Workspace root is resolved through symlink evaluation.
- User paths are required to be relative.
- Existing paths and parent paths are checked for symlink escapes.
- Automated tests cover workspace path validation.

Example usage:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"readFile","path":"README.md"}'
```

## Security Model

The workspace root is the primary security boundary.

All user-provided paths must be resolved through the security layer before filesystem operations are executed.

The daemon must reject:

- Empty paths
- Absolute paths
- Path traversal attempts
- Paths escaping the workspace through symlinks
- Future unsafe write targets
- Future unsafe shell commands

Examples of blocked paths:

```text
/etc/passwd
/home/user/.ssh/id_rsa
../../.env
../../.config
C:\Users\user\.ssh\id_rsa
\\server\share\secret.txt
```

## Architecture Documentation

The project now includes:

```text
docs/ARCHITECTURE.md
```

This document records the main architectural decisions:

- Keep the current prototype simple and CLI-based.
- Separate responsibilities by domain before creating platform-specific abstractions.
- Prepare for Linux, macOS, and Windows without overengineering early.
- Keep the security layer shared across all future transports.
- Avoid unrestricted shell execution.
- Prefer explicit tools such as `gitStatus`, `gitDiff`, `goTest`, and `goFmt`.
- Add cross-platform CI later for Linux, macOS, and Windows.

## Cross-Platform Strategy

Gem Bridge should support Linux, macOS, and Windows.

The current focus is to keep the core logic portable while isolating platform-specific behavior only when necessary.

Cross-platform concerns to handle carefully:

- Filesystem path behavior
- Windows absolute paths and UNC paths
- Symlink handling
- Local configuration directories through `os.UserConfigDir()`
- Native Messaging host registration
- Controlled command execution without exposing arbitrary shell access

Avoid broad premature abstractions such as a generic `OSAdapter` until real platform-specific needs appear.

## Current Git History

Confirmed relevant commits:

```text
docs: add architecture overview
Merge pull request #1 from MTaranto/feat/configurable-workspace
docs: add project context
feat: add configurable workspace root
feat: initialize secure local bridge
```

## Immediate Next Steps

The next recommended steps are:

1. Add `docs/SECURITY_MODEL.md`.
2. Document filesystem threat model and safety rules.
3. Define safe file writing rules before implementing write support.
4. Add safe file writing with strict tests.
5. Add Git tools in small increments:
   - `gitStatus`
   - `gitDiff`
   - later: controlled commit support

## Near-Term Roadmap

Recommended next features, in order:

1. Add `docs/SECURITY_MODEL.md`.
2. Add safe file writing with strict rules.
3. Add Git tools:
   - `gitStatus`
   - `gitDiff`
   - later: controlled commit support
4. Add structured request and response packages.
5. Add WebSocket transport for local development.
6. Add browser extension prototype.
7. Add Native Messaging support.
8. Add approval flow for sensitive operations.
9. Add structured logging.
10. Add cross-platform CI for Linux, macOS, and Windows.

## Long-Term Vision

Gem Bridge should become a secure local bridge that lets browser-based AI assistants interact with local development projects without requiring full IDE migration or unsafe filesystem exposure.

The ideal final architecture is:

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

Future tool categories may include:

- Filesystem tools
- Git tools
- Test/build tools
- Search tools
- Terminal tools with strict allow/block rules
- Browser extension messaging
- Native Messaging integration
- Approval prompts for risky actions

## Important Preferences

The user prefers:

- Explanations in Portuguese.
- Code comments in English.
- Go code that is idiomatic, modular, and professional.
- Direct but educational guidance.
- Incremental development with clean Git history.
- Feature branches for code changes and larger documentation updates.
- Conventional Commits with lowercase descriptions.
- Validation before commits.
- Full-file replacements when that reduces syntax or continuity errors.
- Using Git as the source of truth if the chat becomes long or context is uncertain.
- Avoiding unnecessary command-style boxes in explanations; reserve code blocks mainly for commands, code, filenames, commit messages, or exact outputs.

## Continuity Rule

If there is uncertainty about the current code state, ask the user to provide one of the following before continuing:

```bash
git status
git diff
git log --oneline --decorate --graph --all
```

For file-specific uncertainty, ask for:

```bash
sed -n '1,220p' path/to/file.go
```

Do not ask for file content unnecessarily when the file is already available in the project sources and the local Git state is confirmed clean and aligned with `origin/main`.

Do not guess the current local code state when Git output or file contents are needed.
