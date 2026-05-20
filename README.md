# ⚡ HACKLAB

> Your terminal hacking playground.

Spin up vulnerable lab environments with one command, complete guided objectives, and track your progress — all from the terminal.

## Requirements

- **Go 1.21+** (to build from source)
- **Docker** (to run lab containers)
- **Docker Compose** (usually bundled with Docker Desktop)

## Install

### From source

```bash
# Clone the repo
git clone https://github.com/<you>/hacklab.git
cd hacklab

# Build
go build -o hacklab .

# Install globally (optional)
sudo mv hacklab /usr/local/bin/

# Or add the current directory to your PATH
echo 'export PATH="$HOME/Projects/hacklab:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Go install (if pushed to a public repo)

```bash
go install github.com/<you>/hacklab@latest
```

### Verify the install

```bash
hacklab --version
hacklab --help
```

## Quick Start

```bash
# List available labs
hacklab list

# Start a lab (spins up Docker + interactive TUI)
hacklab start juice-shop

# Stop a lab
hacklab stop juice-shop

# Check what's running
hacklab status

# Add a lab from git
hacklab add https://github.com/user/hacklab-sqli-lab

# Remove a lab
hacklab remove juice-shop
```

## Lab Format

Labs are directories with a `lab.yml` manifest. They're designed to be hackable — just create a folder, write a YAML file, and you've got a lab.

### Single Container Lab

```yaml
name: OWASP Juice Shop
version: "1.0"
description: "Find and exploit 5 vulnerabilities"
difficulty: beginner
author: you
tags: [web, owasp, sqli]

image: bkimminich/juice-shop:latest
port: 3000

wait_for: "http://localhost:3000"  # readiness check
wait_secs: 45

objectives:
  - name: "Bypass login with SQL injection"
    category: "injection"
    hints:
      - "Try the email field"
      - "classic: ' OR 1=1 --"
    flag: "flag{sql_injection_master}"

  - name: "Steal admin JWT token"
    category: "auth"
    hints:
      - "Check the JWT secret"
    flag: "flag{jwt_broken}"
```

### Multi-Container Lab (docker-compose)

```yaml
name: SQL Injection Lab
description: "Practice SQLi against PHP + MySQL"
difficulty: beginner

compose_file: docker-compose.yml
wait_for: "http://localhost:8080"
wait_secs: 30

objectives:
  - name: "Bypass login authentication"
    category: "sqli"
    hints:
      - "Try: admin' -- "
    flag: "flag{sqli_login_bypass}"
```

Then drop a `docker-compose.yml` in the same directory.

## Interactive TUI

When you start a lab, you get an interactive terminal session:

```

██╗  ██╗ █████╗  ██████╗██╗  ██╗██╗      █████╗ ██████╗ 
██║  ██║██╔══██╗██╔════╝██║ ██╔╝██║     ██╔══██╗██╔══██╗
███████║███████║██║     █████╔╝ ██║     ███████║██████╔╝
██╔══██║██╔══██║██║     ██╔═██╗ ██║     ██╔══██║██╔══██╗
██║  ██║██║  ██║╚██████╗██║  ██╗███████╗██║  ██║██████╔╝
╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═════╝ 

  📡 Target: http://localhost:3000

  OBJECTIVES
    ◻  1. Bypass login with SQL injection  [injection]
    ◻  2. Steal admin JWT token            [auth]
    ◻  3. Access admin panel               [broken-auth]
    ◻  4. Find forgotten backup file        [sensitive-data]
    ◻  5. Exploit directory traversal       [injection]

  ─────────────────────────────────

  ❯ submit flag{sql_injection_master}

  ✅ Correct! 'Bypass login with SQL injection' completed
```

**Commands:**
- `objectives` — show all objectives
- `hint N` — get a hint for objective N
- `submit FLAG` — submit a flag
- `url` — show the target URL
- `quit` — exit

## Storage

Everything lives in `~/.hacklab/`:

```
~/.hacklab/
├── labs/           ← your labs (just folders + lab.yml)
│   ├── juice-shop/
│   └── sqli-lab/
└── progress.json   ← local progress tracking
```

## Why?

- TryHackMe/HTB = web UI, accounts, subscriptions
- VulnHub = manual VM setup, no guidance
- Hacklab = one command, guided challenges, local-first, no accounts

## Extending

Since labs are just folders with YAML:
- Write your own labs and share them
- Fork repos, modify flags, add objectives
- Point at any Docker image or compose setup
- No lock-in — it's all local files and Docker
