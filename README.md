# Freezo

A simple, self-hosted application to track food inventory in your freezers. Built with Go, Flutter, and SQLite, designed for homelab environments.

## Architecture

- **Backend**: Go (Golang) with Chi router and SQLite database.
- **Frontend**: React (Vite).
- **Deployment**: Kubernetes via Helm Chart.

## Prerequisites

- Go 1.23+
- Node.js 20+
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

### Local Testing (Docker)

To test the full Docker build locally:

```bash
make run-docker
```
Access the app at `http://localhost:5000`.
Stop the containers with `make stop-docker`.

## Deployment

### Build and Push (Unified Versioning)

The project uses a `Makefile` to build and push Docker images tagged with the current Git version.

1.  **Tag the release**:
    ```bash
    git tag v0.1.0
    ```

2.  **Build and Push**:
    ```bash
    # Build both backend and frontend images
    make build

    # Push to registry
    make push
    ```

    This will create images like `freezo-backend:v0.1.0` and `freezo-frontend:v0.1.0`.

### Deploy with Helm

```bash
helm install freezo ./charts/freezo
```

## Features

- Manage multiple Freezers/Locations.
- Track Items with quantity and expiry dates.
- Categorize items.
- Responsive UI for Desktop and Mobile.

---

This project was vibe coded with [Google Antigravity](https://antigravity.google/) and Gemini 3 Pro.
