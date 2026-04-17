# Security Policy

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Instead, email **sean@lidy.me** with:

- A description of the vulnerability and its potential impact
- Steps to reproduce or a minimal proof-of-concept
- Any suggested mitigations if you have them

You can expect an acknowledgement within 48 hours and a resolution timeline within 7 days depending on severity.

## Scope

gowall reads image files from disk, writes config files to user directories, and reads environment variables and wallpaper manager state to detect the current wallpaper. Security issues are most likely to involve path traversal or unsafe file writes when processing untrusted image paths or template output paths. These are treated as bugs with high priority.
