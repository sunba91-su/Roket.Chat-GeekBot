# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | ✅ |

## Reporting a Vulnerability

If you discover a security vulnerability, please disclose it responsibly by emailing the maintainer directly. **Do not** open a public GitHub issue.

Please include:
- A clear description of the issue
- Steps to reproduce
- Impact assessment
- Any suggested fix (if applicable)

You should receive a response within 48 hours. If the issue is confirmed, a fix will be released as soon as possible — typically within 7 days.

## Best Practices for Deployment

- Store `ROCKETCHAT_BOT_PASSWORD` and other secrets in a secure vault or Docker secrets — never commit them to the repository.
- Create a dedicated Rocket.Chat bot user with the minimum permissions needed (`bot` role).
- Restrict the bot's database file to the bot user only (`chmod 600`).
- Run the container with `--read-only` root filesystem when not using SQLite on ephemeral storage.
- Keep the Docker image and base dependencies up to date.
