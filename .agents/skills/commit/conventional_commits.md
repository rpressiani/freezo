# Conventional Commits Specification (v1.0.0)

The Conventional Commits specification is a lightweight convention on top of commit messages. It provides an easy set of rules for creating an explicit commit history, which makes it easier to write automated tools on top of.

---

## Structural Specification

```
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

### 1. Header Format

- **Type**: MUST be one of the following:
  - `feat`: A new feature
  - `fix`: A bug fix
  - `docs`: Documentation only changes
  - `style`: Code style/formatting changes (whitespace, formatting, missing semi-colons, etc.)
  - `refactor`: A code change that neither fixes a bug nor adds a feature
  - `perf`: A code change that improves performance
  - `test`: Adding missing tests or correcting existing tests
  - `build`: Changes that affect the build system or external dependencies
  - `ci`: Changes to CI configuration files and scripts
  - `chore`: Other changes that don't modify src or test files
  - `revert`: Reverts a previous commit

- **Scope**: OPTIONAL noun describing a section of the codebase surrounded by parentheses, e.g., `fix(parser):`, `feat(auth):`, `chore(deps):`.

- **Breaking Change Indicator**: OPTIONAL `!` placed immediately before the `:` to indicate a breaking change, e.g., `feat(api)!: change payload schema`.

- **Description**: REQUIRED short summary of code changes immediately following the colon and space.
  - Use imperative, present tense ("add", "fix", "change", NOT "added", "fixes", "changed").
  - Do NOT capitalize the first letter unless it is a proper noun or code symbol.
  - Do NOT end with a period `.`.
  - Limit total header line to 72 characters.

### 2. Body Format

- OPTIONAL longer description providing contextual details about code changes (the *what* and *why*, not the *how*).
- MUST begin one blank line after the description header.
- Can consist of multiple paragraphs or bulleted lists.

### 3. Footer Format

- OPTIONAL one or more footers, starting one blank line after the body (or header if body is omitted).
- Each footer MUST consist of a word token, followed by either a `:<space>` or `<space>#` separator, followed by a string value.
- Token MUST use `-` in place of whitespace, e.g., `Reviewed-by:`, `Refs:`, `Closes:`.
- Exception: `BREAKING CHANGE:` or `BREAKING-CHANGE:` MAY be used as tokens.
- `BREAKING CHANGE:` footer MUST consist of the uppercase text `BREAKING CHANGE:`, followed by a space and description of what broke and migration instructions.

---

## Examples

### Feature with scope
```
feat(lang): add Polish language support
```

### Bug fix with scope and body
```
fix(auth): resolve JWT expiration validation bypass

Ensure token expiration timestamp is checked against server time before
granting access to protected endpoints.
```

### Breaking change with `!` and footer
```
feat(api)!: remove deprecated v1 user endpoint

BREAKING CHANGE: The /v1/users endpoint has been removed. Use /v2/users instead.
```

### Chore commit
```
chore(deps): update dependency helm to v3.12.0
```
