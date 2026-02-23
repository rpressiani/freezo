# Contributing to Freezo

First off, thank you for considering contributing to Freezo! It's people like you that make self-hosted and open-source software such a great community to participate in.

## Where do I go from here?

If you've noticed a bug or have a feature request, make sure to check if there's already an [Issue](https://github.com/rpressiani/freezo/issues) open for it. If not, go ahead and open a new one!

If you want to contribute code, please **open an issue first** to discuss the proposed change before putting in a lot of effort. This helps ensure your work aligns with the project's goals and avoids duplicate efforts.

## Local Development Setup

Freezo consists of a Go backend and a React (Vite) frontend.

### Prerequisites

- [Go](https://golang.org/doc/install) (1.25 or newer)
- [Node.js](https://nodejs.org/) (22 or newer)
- Make (optional, but convenient for running tasks)

### Running the Application Locally

The easiest way to run the application locally is using the provided `Makefile` in the root of the repository.

You will need two terminal windows:

**Terminal 1 (Backend):**
```bash
make run-backend
```
The backend API will start on `http://localhost:8080`. The database `freezer.db` will be created automatically in the backend directory.

**Terminal 2 (Frontend):**
```bash
make run-frontend
```
The frontend Vite development server will start on `http://localhost:5173` with Hot Module Replacement (HMR) enabled.

## Code Style and Guidelines

### Backend (Go)
- We follow standard Go formatting. Please run `go fmt ./...` before committing.
- Ensure your code passes all tests: `make test` or `go test ./...` in the `backend` directory.
- Keep handlers small and push business logic to models or dedicated service layers where appropriate.
- When creating new features that touch the database, remember to write tests and ensure they clean up any temporary files they create.

### Frontend (React/TypeScript)
- We use Prettier for code formatting. Please ensure your code is formatted correctly.
- We use Tailwind CSS for styling. Try to stick to the existing design patterns.
- Ensure all new components are typed with TypeScript. Avoid using `any` if possible.

## Submitting a Pull Request

1. **Fork** the repository.
2. **Clone** it to your local machine.
3. **Create a branch** for your feature or bug fix: `git checkout -b feature/your-feature-name` or `fix/your-bug-fix-name`.
4. **Make your changes** and test them thoroughly.
5. **Commit your changes**: Try to write clear, descriptive commit messages.
6. **Push** your branch to your fork: `git push origin feature/your-feature-name`.
7. **Open a Pull Request** against the `main` branch of the original repository.

Please provide a clear description of what the PR does and link to any relevant issues.

## Helm Chart

If you are modifying the Helm chart (`charts/freezo`), please make sure to bump the `version` (and `appVersion` if applicable) in `charts/freezo/Chart.yaml` before submitting your PR. Helm chart releases are automated by GitHub Actions when a change to the `charts/` directory is pushed to the `main` branch.

## Questions?

If you have any questions, feel free to open a Discussion or an Issue! We are happy to help you get started.
