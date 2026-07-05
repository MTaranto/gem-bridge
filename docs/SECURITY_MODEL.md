# Security Model

[Leia em português brasileiro](./SECURITY_MODEL.pt-br.md)

## Purpose

This document defines the security model for Gem Bridge.

Gem Bridge is a local-first daemon that exposes controlled development tools to browser-based AI assistants. Its main security goal is to allow useful local automation without giving the assistant unrestricted access to the user's machine.

The project must treat every request from an AI client, browser extension, script, or future transport as untrusted input.

## Core Security Principle

The configured workspace root is the main security boundary.

Gem Bridge may operate only inside the authorized workspace. Any request that attempts to access files, directories, commands, or Git operations outside this boundary must be rejected before the operation runs.

This rule must apply equally to all current and future transports:

- CLI
- WebSocket
- Browser extension
- Native Messaging
- Any future local protocol

Security rules must live in shared internal packages, not inside a specific transport layer.

## Trust Boundaries

Gem Bridge has the following trust boundaries:

1. **AI client boundary**
   - Requests coming from an AI assistant are untrusted.
   - Tool names, paths, arguments, and command parameters must be validated.

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

Allowed paths must resolve to a location inside the authorized workspace after cleaning, joining, and symlink evaluation.

## Symlink Safety

Symlinks are security-sensitive because a path may look like it is inside the workspace while actually pointing outside it.

Gem Bridge must block symlink escapes by validating existing paths and parent paths before filesystem operations run.

Example dangerous layout:

```text
workspace/
  link-to-home -> /home/user
```

A request such as this must be rejected:

```json
{
  "tool": "readFile",
  "path": "link-to-home/.ssh/id_rsa"
}
```

The security layer must ensure the final resolved path remains inside the workspace root.

## Read Operations

Read operations are the safest initial filesystem capability, but they are still sensitive.

Read tools must:

- Require a relative path
- Resolve the path through the workspace security layer
- Refuse access outside the workspace
- Return structured JSON errors
- Avoid leaking host-specific filesystem details when possible

Current read-oriented tools:

- `listDirectory`
- `readFile`

Future read tools should follow the same security path before touching the disk.

## Write Operations

Write operations are more dangerous than read operations and must be implemented with stricter rules.

Future write tools must:

- Require relative paths
- Resolve paths through the workspace security layer
- Block traversal and symlink escapes
- Refuse writes outside the workspace
- Avoid overwriting existing files unless explicitly allowed
- Consider file size limits
- Return structured JSON errors
- Be covered by automated tests
- Eventually support user approval for sensitive writes

Initial file writing should be narrow and explicit. Avoid broad tools that can modify arbitrary files without clear rules.

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

Each command tool must:

- Call a known executable
- Use controlled arguments
- Run inside the authorized workspace
- Avoid shell interpolation
- Avoid destructive behavior by default
- Return structured output
- Be tested where practical

Dangerous commands and shell features must remain blocked unless a future approval model explicitly supports them.

## Git Operation Rules

Git tools should be introduced incrementally.

Safer initial Git tools:

- `gitStatus`
- `gitDiff`

These are read-oriented and useful for validation.

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

Future operations that may require approval include:

- Writing files
- Overwriting files
- Deleting files
- Running tests or build commands with side effects
- Staging files
- Creating commits
- Switching branches
- Merging branches
- Pushing to remotes
- Installing dependencies
- Running any tool that may modify the local system

Approval should be explicit, auditable, and understandable to the user.

## Error Handling

Security failures must return structured errors.

A failed request should not reveal unnecessary host details.

In CLI mode, returning a non-zero exit status for failed tool requests is acceptable.

In future daemon mode, a failed tool request must not crash the process. The daemon should return a structured JSON error and keep running.

## Cross-Platform Security Considerations

Path safety must be validated across Linux, macOS, and Windows.

Special care is required for:

- Unix absolute paths
- Windows drive-letter paths
- Windows UNC paths
- Mixed path separators
- Symlink and junction behavior
- Case sensitivity differences
- Config directory locations
- Native Messaging host registration paths

Cross-platform CI should eventually run security tests on:

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
- Write tools without clear safety rules
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
- Future write safety rules
- Future command allowlist behavior
- Future Git tool validation

Security tests should be small, explicit, and easy to understand.

## Final Rule

When in doubt, Gem Bridge must choose the safer behavior.

A blocked safe request is an inconvenience. An allowed unsafe request can expose private files, damage a project, or execute unwanted local actions.
