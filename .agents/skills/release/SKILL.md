---
name: release
description: Release automation prompt command `/release` for Freezo. Calculates SemVer version, generates changelog, updates Helm chart, presents a detailed plan for developer confirmation, tags git commit, creates GitHub release, and republishes the Helm chart.
---

# Freezo Release Skill (`/release`)

This skill automates the end-to-end release process for Freezo whenever the prompt command `/release` is invoked or when asked to create a release.

---

## Workflow Overview

When `/release` is invoked, execute the release workflow in two distinct phases:

1. **Planning Phase**: Analyze git history, compute SemVer bump, generate changelog, determine Helm chart version updates, format a **Release Plan**, and **ASK FOR DEVELOPER CONFIRMATION**.
2. **Execution Phase**: (Only after explicit user approval) Update Helm chart, commit, tag, push to GitHub, create GitHub release, and verify Helm chart publishing.

---

## Phase 1: Planning & Analysis

1. **Gather Context**:
   - Run the helper script if available:
     ```bash
     python3 .agents/skills/release/scripts/release_helper.py
     ```
     Or manually inspect git tags and history:
     ```bash
     git describe --tags --abbrev=0
     git log <latest-tag>..HEAD --pretty=format:"%h %s"
     ```

2. **Determine Version Bump (SemVer)**:
   - **MAJOR** (`v(X+1).0.0`): Any commit containing `BREAKING CHANGE:` in body/subject or `!:` prefix (e.g. `feat!:`, `fix!:`).
   - **MINOR** (`vX.(Y+1).0`): Any commit starting with `feat` or `feat(...)`.
   - **PATCH** (`vX.Y.(Z+1)`): Any commits starting with `fix`, `refactor`, `perf`, `docs`, `chore`, `ci`, etc.

3. **Determine Helm Chart Updates**:
   - Inspect `charts/freezo/Chart.yaml`.
   - Update `appVersion` to match the new release version (e.g. `"v0.3.0"`).
   - Increment `version` patch number (e.g. `0.1.5` -> `0.1.6`).

4. **Generate Categorized Changelog**:
   Group commits since the last release tag into:
   - 🚀 **Features** (`feat`)
   - 🐛 **Bug Fixes** (`fix`)
   - 🧰 **Maintenance & Refactoring** (`chore`, `refactor`, `build`, `ci`, `docs`, `perf`)
   - ⚠️ **Breaking Changes** (if any)

5. **Generate Release Plan & Request Confirmation**:
   Present the plan to the developer in the following format:

   ```markdown
   # 🚀 Release Plan: Freezo <PROPOSED_VERSION>

   ## Overview
   - **Current Tag**: `<CURRENT_TAG>`
   - **Proposed Version**: `<NEW_VERSION>` (SemVer Bump: `<BUMP_TYPE>`)
   - **Helm Chart Update**: `charts/freezo/Chart.yaml`
     - `version`: `<OLD_CHART_VERSION>` ➔ `<NEW_CHART_VERSION>`
     - `appVersion`: `<OLD_APP_VERSION>` ➔ `<NEW_APP_VERSION>`

   ## 📝 Changelog Preview
   <GENERATED_CHANGELOG>

   ## ⚡ Proposed Actions
   1. **Update Helm Chart**: Update `charts/freezo/Chart.yaml` (`version` & `appVersion`).
   2. **Commit Changes**: Create git commit `chore(release): prepare release <NEW_VERSION>`.
   3. **Tag Commit**: Create annotated git tag `<NEW_VERSION>`.
   4. **Push to Remote**: `git push origin main` & `git push origin <NEW_VERSION>`.
   5. **GitHub Release**: Create GitHub release using `gh release create <NEW_VERSION>`.
   6. **Republish Helm Chart**: Push triggers GitHub Action `.github/workflows/chart-release.yml` to publish updated Helm chart to `gh-pages`.

   ---
   **Developer Approval Required**: Do you confirm proceeding with this release?
   ```

   > [!IMPORTANT]
   > **STOP HERE**. Do NOT run any modifying commands (`git commit`, `git tag`, `git push`, `gh release create`, or file edits) until the developer explicitly confirms!

---

## Phase 2: Execution (Post-Confirmation)

Upon developer confirmation, execute the following steps in sequence:

1. **Update Helm Chart (`charts/freezo/Chart.yaml`)**:
   ```bash
   python3 .agents/skills/release/scripts/release_helper.py apply
   ```
   Or manually update `version` and `appVersion` in `charts/freezo/Chart.yaml`.

2. **Commit Version Update**:
   ```bash
   git add charts/freezo/Chart.yaml
   git commit -m "chore(release): prepare release <NEW_VERSION>"
   ```

3. **Tag Git Commit**:
   ```bash
   git tag -a <NEW_VERSION> -m "Release <NEW_VERSION>"
   ```

4. **Push Commit and Tag to GitHub**:
   ```bash
   git push origin main
   git push origin <NEW_VERSION>
   ```

5. **Create GitHub Release**:
   ```bash
   gh release create <NEW_VERSION> \
     --title "<NEW_VERSION>" \
     --notes "<CHANGELOG_CONTENT>"
   ```

6. **Verify Helm Chart Republishing**:
   - Note that pushing to `main` with changes under `charts/` automatically triggers `.github/workflows/chart-release.yml`.
   - Verify workflow dispatch/status using:
     ```bash
     gh run list --workflow=chart-release.yml --limit 1
     ```

7. **Report Completion**:
   Summarize the completed release with links to the GitHub Release and updated Helm Chart.
