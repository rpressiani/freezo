# Freezo

A simple, self-hosted application to track food inventory in your freezers. Built with Go, React, and SQLite, designed for homelab environments.

## Architecture

- **Backend**: Go (Golang) with Chi router and SQLite database.
- **Frontend**: React (Vite).
- **Deployment**: Kubernetes via Helm Chart.

## Prerequisites

- Go 1.25+
- Node.js 22+
- Docker
- Kubernetes Cluster (e.g., k3s, minikube)
- Helm

## Local Development

### Local Development (Source)

You can run the backend and frontend in separate terminals:

**Terminal 1 (Backend):**
```bash
make run-backend
```

**Terminal 2 (Frontend):**
```bash
make run-frontend
```

## Release Process

### Application Release
To release a new version of the application (Frontend + Backend):
1.  **Tag the release**:
    ```bash
    git tag v0.1.0
    git push origin v0.1.0
    ```
    This triggers the `release.yml` workflow which builds and pushes the Docker images.

### Helm Chart Release
To release a new version of the Helm chart:
1.  **Update Config**: Bump the version in `charts/freezo/Chart.yaml`.
2.  **Push**: Commit and push the change to `main`.
    This triggers the `chart-release.yml` workflow which packages and releases the chart to GitHub Pages.

## Deployment on Kubernetes

```bash
helm repo add freezo https://rpressiani.github.io/freezo
helm install freezo freezo/freezo
```

## Features

- Manage multiple Freezers/Locations.
- Track Items with quantity and expiry dates.
- Categorize items.
- Responsive UI for Desktop and Mobile.

---

This project was vibe coded with [Google Antigravity](https://antigravity.google/) and Gemini 3 Pro.
