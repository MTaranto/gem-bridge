# Security Model

[Leia em português brasileiro](./SECURITY_MODEL.pt-br.md)

## Purpose

This document defines the security model for Gem Bridge.

Gem Bridge is a local-first daemon that exposes controlled development tools to browser-based AI assistants. Its main security goal is to allow useful local automation without giving the assistant unrestricted access to the user's machine.

The project must treat every request from an AI client, browser extension, script, or future transport as untrusted input.

## Core Security Principle

The configured workspace root is the main security boundary.

Gem Bridge may operate only inside the authorized workspace. Any request that attempts to access files, directories, commands, or Git operations outside this boundary must be rejected before the operation runs.

Security rules must live in shared internal packages, not inside a specific transport layer.

## Trust Boundaries

1. **AI client boundary**
   - Requests coming from an AI assistant are untrusted.
   - Tool names, paths, arguments, content payloads, and command parameters must be validated.

2. **Transport boundary**
   - CLI, WebSocket, and Native Messaging are only delivery mechanisms.
   - A trusted transport does not make the request itself trusted.

3. **Workspace boundary**
   - The workspace root is the only authorized filesystem area.
   - All filesystem paths must be resolved through the security layer.

4. **Command boundary**
   - Local command execution is dangerous.
   - Gem Bridge must never expose unrestricted shell access.

5. **Git boundary**
   - Git operations can modify project history, branches, and remote repositories.
   - Read-only Git tools are safer than write operations.
   - Sensitive Git operations require additional validation and may require user approval later.

## Filesystem Security Rules

All user-provided paths must be relative paths.

Gem Bridge must reject:

- Empty paths
- Absolute paths
- Path traversal attempts
- Paths that escape the workspace through symlinks
- Platform-specific absolute paths
- UNC paths on Windows
- Any path that cannot be safely resolved inside the workspace

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

Allowed paths must resolve to a location inside the authorized workspace after cleaning, joining, separator normalization, and symlink evaluation.

## Symlink Safety

Symlinks are security-sensitive because a path may look like it is inside the workspace while actually pointing outside it.

Gem Bridge must block symlink escapes by validating existing paths and parent paths before filesystem operations run.

## Read Operations

Current read-oriented tools:

- `listDirectory`
- `readFile`

Read tools must require relative paths, resolve paths through the workspace security layer, refuse access outside the workspace, and return structured JSON errors.

## Write Operations

Current write-oriented tools:

- `writeFile`

The current `writeFile` implementation is intentionally conservative. It:

- Requires a relative path
- Resolves the path through the workspace security layer
- Blocks traversal and symlink escapes
- Refuses writes outside the workspace
- Refuses to overwrite existing files
- Enforces a maximum content size
- Returns structured JSON errors
- Is covered by automated tests

This first version creates new text files only. Overwriting, patching, appending, deleting, or renaming files should remain separate future tools with explicit safety rules.

Future write capabilities must avoid broad tools that can modify arbitrary files without clear rules.

## Command Execution Rules

Gem Bridge must not expose arbitrary shell execution.

The daemon must not accept requests such as:

```json
{
  "tool": "runCommand",
  "command": "rm -rf ~/project"
}
```

Instead, command-like behavior must be exposed through explicit tools with controlled arguments.

Safer examples:

- `goFmt`
- `goTest`
- `gitStatus`
- `gitDiff`

Each command tool must call a known executable, use controlled arguments, run inside the authorized workspace, avoid shell interpolation, avoid destructive behavior by default, return structured output, and be tested where practical.

## Git Operation Rules

Git tools should be introduced incrementally.

Safer initial Git tools:

- `gitStatus`
- `gitDiff`

More sensitive Git tools:

- `gitAdd`
- `gitCommit`
- `gitCheckout`
- `gitMerge`
- `gitPush`

Sensitive Git tools can change history, branches, remote state, or staged content. They should require stricter validation and may require user approval in a future version.

Git tools must run only inside the authorized workspace.

## Approval Flow

Gem Bridge does not currently implement an approval flow, but the architecture should prepare for it.

Future operations that may require approval include overwriting files, deleting files, running commands with side effects, staging files, creating commits, switching branches, merging, pushing, installing dependencies, or running broader system-modifying tools.

Approval should be explicit, auditable, and understandable to the user.

## Error Handling

Security failures must return structured errors.

In CLI mode, returning a non-zero exit status for failed tool requests is acceptable.

In future daemon mode, a failed tool request must not crash the process. The daemon should return a structured JSON error and keep running.

## Cross-Platform Security Considerations

Path safety must be validated across Linux, macOS, and Windows.

Special care is required for Unix absolute paths, Windows drive-letter paths, Windows UNC paths, mixed path separators, symlink and junction behavior, case sensitivity differences, config directory locations, and Native Messaging host registration paths.

Cross-platform CI currently runs Go formatting and tests on:

```text
ubuntu-latest
macos-latest
windows-latest
```

## Non-Goals

Gem Bridge should not support these behaviors at the current stage:

- Arbitrary shell execution
- Unrestricted filesystem access
- Remote daemon exposure
- GUI automation
- Silent destructive operations
- Sensitive Git operations without validation
- Broad write tools without clear safety rules
- Platform-specific shortcuts that bypass the shared security layer

## Testing Requirements

Security behavior must be tested before features are considered complete.

Current and future tests should cover:

- Workspace root creation
- Relative path resolution
- Empty path rejection
- Absolute path rejection
- Path traversal rejection
- Symlink escape rejection
- Read access inside the workspace
- Read access outside the workspace
- Write creation inside the workspace
- Write overwrite rejection
- Write size limit rejection
- Write rejection for unsafe paths
- Write rejection for symlink parent escapes
- Future command allowlist behavior
- Future Git tool validation

## Final Rule

When in doubt, Gem Bridge must choose the safer behavior.

A blocked safe request is an inconvenience. An allowed unsafe request can expose private files, damage a project, or execute unwanted local actions.
