---
name: fix-dependabot
description: Framework-agnostic skill to inventory, consolidate, test, and resolve open Dependabot pull requests across any repository (Go, Node, Python, Rust, Docker, Helm, GitHub Actions).
---

# Fix Dependabot Skill (`/fix-dependabot`)

Use this skill to systematically resolve open Dependabot pull requests in **any codebase or ecosystem** by consolidating dependency updates, performing local verification, pushing a clean commit to `main`, and allowing Dependabot to auto-close satisfied PRs.

---

## Workflow Steps

### Step 1: Inventory & Categorize Open Dependabot PRs

1. **List all open PRs**:
   ```bash
   gh pr list --state open
   ```

2. **Check CI status for each Dependabot PR**:
   ```bash
   for num in $(gh pr list --state open --json number -q '.[].number'); do
     echo "=== PR #$num ==="
     gh pr checks $num || true
   done
   ```

3. **Group open PRs by Ecosystem**:
   - **Node.js / NPM / Yarn / pnpm**: `package.json`, lockfiles
   - **Go**: `go.mod`, `go.sum`
   - **Python**: `pyproject.toml`, `requirements.txt`, `poetry.lock`
   - **Rust**: `Cargo.toml`, `Cargo.lock`
   - **Helm / Docker / Kubernetes**: `Chart.yaml`, `Dockerfile`
   - **GitHub Actions Workflows**: `.github/workflows/*.yml`

---

### Step 2: Consolidate Dependencies & Test Locally

1. **Update Ecosystem Dependencies**:
   - **Go**: `cd <backend> && go get <pkg>@<ver> && go mod tidy`
   - **Node.js**: `cd <frontend> && npm install <pkg1>@<ver> <pkg2>@<ver>` (batch install to resolve peer dependencies cleanly)
   - **Python**: `poetry update <pkg>` or `pip install -U <pkg>`
   - **Rust**: `cargo update -p <pkg> --precise <ver>`
   - **GitHub Actions**: Update `uses: owner/action@version` in `.github/workflows/*.yml`.

2. **Verify GitHub Action Release Tags**:
   Before updating action versions in workflow files, verify tag naming conventions on GitHub (check for leading `v` prefixes):
   ```bash
   gh release list --repo <owner/action> --limit 5
   ```
   *Example*: `aquasecurity/trivy-action` uses `v0.36.0` (with leading `v`), whereas `0.35.0` existed without `v`.

3. **Execute Project Verification Suite**:
   Run the repository's native test and build commands:
   - Run unit/integration tests (e.g., `go test ./...`, `npm test`, `pytest`, `cargo test`).
   - Run build scripts (e.g., `npm run build`, `go build ./...`, `cargo build`).
   - Run linter & security audit tools (e.g., `npm audit`, `trivy`, `golangci-lint`).

---

### Step 3: Commit and Push to Main

1. Stage modified manifest, lock, and workflow files:
   ```bash
   git add <manifest_files> <lock_files> .github/workflows/
   ```
2. Commit using Conventional Commits specification:
   ```bash
   git commit -m "chore(deps): update dependencies"
   ```
3. Push commit to primary branch:
   ```bash
   git push origin main
   ```

---

### Step 4: Re-trigger Dependabot & Verify Closure

1. Trigger Dependabot re-evaluation on all open Dependabot PRs:
   ```bash
   for num in <DEPENDABOT_PR_NUMBERS>; do
     gh pr comment $num -b "@dependabot rebase"
   done
   ```
2. **Verify Automatic PR Closure**:
   - Dependabot will re-evaluate each PR against `main`.
   - PRs whose dependencies are satisfied on `main` will automatically close.

---

## Troubleshooting Matrix

| Symptom | Root Cause | Remediation |
| :--- | :--- | :--- |
| **`Unable to resolve action ... unable to find version`** | Missing leading `v` prefix in GitHub Action tag (e.g., `@0.36.0` vs `@v0.36.0`). | Check releases via `gh release list --repo <action>` and format tag correctly. |
| **Peer Dependency Conflicts (`ERESOLVE`)** | Installing packages individually causes version mismatches between peer dependencies. | Batch update all related packages in a single installer command. |
| **CI Failures on Old PR Branches** | PR branch predates recent base image security patches or main branch refactors. | Pushing consolidated updates directly to `main` incorporates latest base fixes. |
