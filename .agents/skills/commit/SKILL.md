---
name: commit
description: Generate conventional git commit messages following the Conventional Commits v1.0.0 specification. Triggered by `/commit` or requests to generate/write commit messages.
---

# Commit Skill (`/commit`)

Generates structured, high-quality git commit messages adhering strictly to the **Conventional Commits v1.0.0** specification based on staged (`git diff --cached`) or working tree (`git diff`) changes.

---

## Workflow Overview

When `/commit` is invoked or when asked to write/generate a commit message:

1. **Check Git Status & Diffs**:
   - Execute `git status` and `git diff --cached` to inspect staged changes.
   - If no changes are staged, inspect unstaged changes with `git diff`. Inform the user that changes are not yet staged, show a summary of changed files, and ask if they would like to stage all or specific files (e.g. `git add .` or `git add <files>`).

2. **Analyze Changes**:
   - Read and analyze the diff output to understand the scope and intent of the changes.
   - Determine the appropriate **Conventional Commit Type**:
     - `feat`: A new feature for the user / codebase.
     - `fix`: A bug fix.
     - `docs`: Documentation changes only.
     - `style`: Formatting, missing semi-colons, whitespace, code styling (no production code change).
     - `refactor`: Code change that neither fixes a bug nor adds a feature.
     - `perf`: Code change that improves performance.
     - `test`: Adding missing tests or correcting existing tests.
     - `build`: Changes that affect the build system or external dependencies (npm, go, cargo, make, helm).
     - `ci`: Changes to CI configuration files and scripts (GitHub Actions, GitLab CI, etc.).
     - `chore`: Other changes that don't modify src or test files (e.g., updating .gitignore, repo configuration).
     - `revert`: Reverts a previous commit.
   - Determine the **Scope** (optional):
     - Identify the specific section or module of the codebase (e.g., `auth`, `api`, `models`, `deps`, `helm`, `ci`, `ui`).
   - Check for **Breaking Changes**:
     - If the change breaks backward compatibility, append a `!` after the type/scope (e.g., `feat(api)!:` or `fix!:`) AND/OR include `BREAKING CHANGE: <description>` in the commit footer.

3. **Format Commit Message**:
   Follow Conventional Commits v1.0.0 format strictly:
   ```
   <type>[optional scope]: <description>

   [optional body]

   [optional footer(s)]
   ```
   - **Header Line**: Max 72 characters, imperative mood ("add feature" not "added feature" or "adds feature"), lowercase start, no trailing period.
   - **Body** (optional/recommended for non-trivial changes): Blank line before body. Provide concise context on *what* was changed and *why*.
   - **Footer** (optional): Blank line before footer. Use standard trailers (e.g. `BREAKING CHANGE: ...`, `Closes #123`, `Refs #456`).

4. **Present Message & Offer Execution**:
   - Present the recommended commit message in a copyable code block.
   - If the user explicitly requested to commit, run `git commit -m "..."` (or multi-line `git commit -F -`).
   - Otherwise, ask the user if they would like you to execute the commit for them.

---

## Detailed Specification Reference

See [conventional_commits.md](file:///.agents/skills/commit/conventional_commits.md) in this skill folder for full rules and edge-case guidance.
