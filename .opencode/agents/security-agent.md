# security / code-reviewer Agent
**Skills:** code-reviewer, architecture-designer, the-fool
**Scope:** deploy/systemd, lib/watch, permissions

- Review Option A: systemd ProtectSystem=strict, NoNewPrivileges, 0750/0640
- Ensure no rootkit hiding (ps/bpftool visible to sudo)
- Tag: blue-team, authorized testing only

