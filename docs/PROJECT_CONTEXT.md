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
- Current feature branch: `feat/configurable-workspace`

## Current Status

The project already has an initial Go structure with:

- `cmd/gem-bridge/main.go`
- `internal/security/workspace.go`
- `internal/tools/files.go`
- `internal/security/workspace_test.go`
- `README.md`
- `.gitignore`
- `go.mod`

The initial commit was created and pushed to GitHub:

```text
feat: initialize secure local bridge
```

A second feature branch was created:

```text
feat/configurable-workspace
```

This branch added support for configurable workspace root handling and automated tests for workspace path validation.

The feature commit was created and pushed:

```text
feat: add configurable workspace root
```

`go test ./...` passed successfully after the configurable workspace changes.

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
   - Features should be developed in branches.
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

Before each commit, validate changes with:

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

In the current CLI version, failed requests may also return exit status `1`. This is expected for CLI behavior. In a future daemon/server mode, failed tool requests should not crash the daemon.

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
```

## Current Feature Branch

Branch:

```text
feat/configurable-workspace
```

Implemented behavior:

- Added `--workspace` CLI flag.
- New default workspace root is `.`.
- Workspace root is resolved to an absolute path.
- Workspace root is resolved through symlink evaluation.
- User paths are still required to be relative.
- Existing paths and parent paths are checked for symlink escapes.
- Added automated tests for workspace path validation.

Validation already performed:

```bash
go test ./...
```

Expected result:

```text
?    github.com/mtaranto/gem-bridge/cmd/gem-bridge      [no test files]
?    github.com/mtaranto/gem-bridge/internal/tools      [no test files]
ok   github.com/mtaranto/gem-bridge/internal/security
```

## Immediate Next Steps

The next recommended steps are:

1. Open a Pull Request on GitHub for `feat/configurable-workspace`.
2. Review the PR description and changed files.
3. Merge the PR into `main`.
4. Pull/update local `main`.
5. Start the next feature branch.

Suggested PR title:

```text
feat: add configurable workspace root
```

Suggested PR description:

```text
Adds support for configuring the workspace root through a CLI flag and introduces automated tests for workspace path resolution and escape prevention.
```

## Near-Term Roadmap

Recommended next features, in order:

1. Add `docs/PROJECT_CONTEXT.md`.
2. Add `docs/SECURITY_MODEL.md`.
3. Add safe file writing with strict rules.
4. Add Git tools:
   - `gitStatus`
   - `gitDiff`
   - later: controlled commit support
5. Add structured request and response packages.
6. Add WebSocket transport for local development.
7. Add browser extension prototype.
8. Add Native Messaging support.
9. Add approval flow for sensitive operations.
10. Add structured logging.

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
- Feature branches.
- Conventional Commits with lowercase descriptions.
- Validation before commits.
- Full-file replacements when that reduces syntax or continuity errors.
- Using Git as the source of truth if the chat becomes long or context is uncertain.

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

Do not guess the current local code state when Git output or file contents are needed.
