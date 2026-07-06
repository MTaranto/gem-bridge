# Project Context

## Project Name

Gem Bridge

## Purpose

Gem Bridge is a local-first AI tooling bridge written in Go.

The goal is to build a secure local daemon that acts as a bridge between browser-based AI assistants and the user's local development environment. The daemon exposes controlled tools for filesystem access, development workflows, Git operations, and future browser integration without giving unrestricted access to the machine.

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
- `internal/tools/files_test.go`
- `.github/workflows/ci.yml`
- `docs/PROJECT_CONTEXT.md`
- `docs/ARCHITECTURE.md`
- `docs/ARCHITECTURE.pt-br.md`
- `docs/SECURITY_MODEL.md`
- `docs/SECURITY_MODEL.pt-br.md`
- `README.md`
- `README.pt-br.md`
- `.gitignore`
- `go.mod`

Confirmed project milestones:

```text
feat: initialize secure local bridge
feat: add configurable workspace root
docs: add project context
docs: add architecture overview
docs: add brazilian portuguese documentation
docs: add security model
ci: add cross-platform go workflow
fix: reject cross-platform absolute paths
feat: add safe file writing
```

Pull Request #1 was opened from `feat/configurable-workspace` into `main`, reviewed, and merged.

Pull Request #2 was opened from `feat/safe-file-writing` into `main`, passed CI, reviewed, and merged.

The latest confirmed local state:

- `main` is aligned with `origin/main`.
- Working tree is clean.
- Latest merge commit on `main`: `Merge pull request #2 from MTaranto/feat/safe-file-writing`.
- The local feature branch `feat/safe-file-writing` was deleted.
- The remote feature branch `origin/feat/safe-file-writing` was pruned.
- Cross-platform CI is passing.

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
   - Windows drive paths and UNC paths must be rejected when received from clients.

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
   - Pull requests and CI should be used for meaningful code changes.

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
feat: add safe file writing
fix: reject cross-platform absolute paths
ci: add cross-platform go workflow
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
- `writeFile`

The current request format for reading is:

```json
{
  "tool": "readFile",
  "path": "go.mod"
}
```

The current request format for writing is:

```json
{
  "tool": "writeFile",
  "path": "notes.txt",
  "content": "hello from gem bridge"
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
- Empty paths are rejected.
- Unix absolute paths are rejected.
- Windows drive-letter paths are rejected.
- Windows UNC paths are rejected.
- Mixed separators are normalized before path resolution.
- Existing paths and parent paths are checked for symlink escapes.
- Automated tests cover workspace path validation.
- Cross-platform CI validates behavior on Linux, macOS, and Windows.

Example usage:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"readFile","path":"README.md"}'
```

## Current File Tool Behavior

Implemented tools:

- `listDirectory`
- `readFile`
- `writeFile`

Current `writeFile` behavior:

- Creates new text files only.
- Requires a relative path.
- Resolves the path through `Workspace.ResolvePath`.
- Refuses writes outside the workspace.
- Refuses to overwrite existing files.
- Enforces a maximum content size.
- Returns structured JSON errors.
- Has automated tests for successful writes, overwrite rejection, size limits, unsafe paths, and symlink parent escapes.

Example usage:

```bash
go run ./cmd/gem-bridge --workspace . '{"tool":"writeFile","path":"notes.txt","content":"hello from gem bridge"}'
```

## Security Model

The workspace root is the primary security boundary.

All user-provided paths must be resolved through the security layer before filesystem operations are executed.

The daemon must reject:

- Empty paths
- Absolute paths
- Path traversal attempts
- Paths escaping the workspace through symlinks
- Windows drive-letter paths
- Windows UNC paths
- Unsafe write targets
- Future unsafe shell commands

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

## Architecture Documentation

The project includes:

```text
docs/ARCHITECTURE.md
docs/ARCHITECTURE.pt-br.md
```

These documents record the main architectural decisions:

- Keep the current prototype simple and CLI-based.
- Separate responsibilities by domain before creating platform-specific abstractions.
- Prepare for Linux, macOS, and Windows without overengineering early.
- Keep the security layer shared across all future transports.
- Avoid unrestricted shell execution.
- Prefer explicit tools such as `gitStatus`, `gitDiff`, `goTest`, and `goFmt`.
- Run cross-platform CI for Linux, macOS, and Windows.

## Security Documentation

The project includes:

```text
docs/SECURITY_MODEL.md
docs/SECURITY_MODEL.pt-br.md
```

These documents define:

- Workspace trust boundary.
- Filesystem security rules.
- Symlink safety.
- Read and write operation rules.
- Command execution rules.
- Git operation rules.
- Approval flow expectations.
- Error handling.
- Cross-platform security considerations.
- Testing requirements.

## Documentation Language Strategy

Public-facing documentation should be available in English and Brazilian Portuguese.

English remains the primary language for GitHub discoverability and international readability. Brazilian Portuguese versions should be maintained alongside the English documents to support learning, review, and portfolio communication in Brazil.

Each public document should include a clear link to its counterpart:

```text
README.md -> README.pt-br.md
README.pt-br.md -> README.md
docs/ARCHITECTURE.md -> docs/ARCHITECTURE.pt-br.md
docs/ARCHITECTURE.pt-br.md -> docs/ARCHITECTURE.md
docs/SECURITY_MODEL.md -> docs/SECURITY_MODEL.pt-br.md
docs/SECURITY_MODEL.pt-br.md -> docs/SECURITY_MODEL.md
```

## Continuous Integration

The project includes GitHub Actions CI:

```text
.github/workflows/ci.yml
```

The workflow runs on pushes and pull requests targeting `main`.

It validates:

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

The first CI run revealed a real Windows path validation issue, which was fixed by rejecting cross-platform absolute path shapes regardless of the operating system running the daemon.

## Current Git History

Confirmed relevant commits:

```text
Merge pull request #2 from MTaranto/feat/safe-file-writing
feat: add safe file writing
fix: reject cross-platform absolute paths
ci: add cross-platform go workflow
docs: add security model
docs: add brazilian portuguese documentation
docs: update project context
docs: add architecture overview
Merge pull request #1 from MTaranto/feat/configurable-workspace
docs: add project context
feat: add configurable workspace root
feat: initialize secure local bridge
```

## Immediate Next Steps

The next recommended steps are:

1. Update public documentation to reflect `writeFile`, CI, and merged PR #2.
2. Open a new ChatGPT thread inside the Gem Bridge project to keep context fresh.
3. In the next thread, continue with Git tools in small increments:
   - `gitStatus`
   - `gitDiff`
4. Keep command execution explicit and allowlisted.
5. Avoid arbitrary shell execution.

## Near-Term Roadmap

Recommended next features, in order:

1. Add Git tools:
   - `gitStatus`
   - `gitDiff`
2. Add structured request and response packages.
3. Add WebSocket transport for local development.
4. Add browser extension prototype.
5. Add Native Messaging support.
6. Add approval flow for sensitive operations.
7. Add structured logging.
8. Expand safe file mutation rules only when necessary:
   - controlled overwrite
   - append
   - patch/update
   - delete, if ever allowed

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

- To be called Careca.
- Explanations in Brazilian Portuguese.
- Code comments in English.
- Go code that is idiomatic, modular, and professional.
- Direct but educational guidance.
- Incremental development with clean Git history.
- Feature branches for code changes and larger documentation updates.
- Conventional Commits with lowercase descriptions.
- Validation before commits.
- Full-file replacements when that reduces syntax or continuity errors.
- Using Git as the source of truth if the chat becomes long or context is uncertain.
- After commit/push, check source code directly on GitHub when code verification is needed.
- With uncommitted local changes, ask for `git status`, `git diff`, or file contents.
- Avoiding unnecessary command-style boxes in explanations; reserve code blocks mainly for commands, code, filenames, commit messages, or exact outputs.

## Continuity Rule

If there is uncertainty about the current code state after a commit/push, check GitHub first.

If there are local uncommitted changes, ask the user to provide one of the following before continuing:

```bash
git status
git diff
git log --oneline --decorate --graph --all
```

For file-specific uncertainty with local uncommitted changes, ask for:

```bash
sed -n '1,220p' path/to/file.go
```

Do not ask for file content unnecessarily when the file is already available on GitHub and the local Git state is confirmed clean and aligned with `origin/main`.

Do not guess the current local code state when Git output or file contents are needed.
