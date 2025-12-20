# Image names
BACKEND_IMAGE := freezo-backend
FRONTEND_IMAGE := freezo-frontend

# Registry (optional, can be overridden)
REGISTRY ?= localhost:5000

# Version from Git
VERSION := $(shell git describe --tags --always --dirty)

.PHONY: all build push

all: build

build: build-backend build-frontend

build-backend:
	docker build -t $(REGISTRY)/$(BACKEND_IMAGE):$(VERSION) -f backend/Dockerfile backend/
	docker tag $(REGISTRY)/$(BACKEND_IMAGE):$(VERSION) $(REGISTRY)/$(BACKEND_IMAGE):latest

build-frontend:
	docker build -t $(REGISTRY)/$(FRONTEND_IMAGE):$(VERSION) -f frontend/Dockerfile frontend/
	docker tag $(REGISTRY)/$(FRONTEND_IMAGE):$(VERSION) $(REGISTRY)/$(FRONTEND_IMAGE):latest

push: push-backend push-frontend

push-backend:
	docker push $(REGISTRY)/$(BACKEND_IMAGE):$(VERSION)
	docker push $(REGISTRY)/$(BACKEND_IMAGE):latest

push-frontend:
	docker push $(REGISTRY)/$(FRONTEND_IMAGE):$(VERSION)
	docker push $(REGISTRY)/$(FRONTEND_IMAGE):latest

version:
	@echo $(VERSION)

run-backend:
	cd backend && go run main.go

run-frontend:
	cd frontend && npm run dev

run-docker: build
	@echo "Starting Backend..."
	docker run -d --rm --name freezo-backend -p 8080:8080 -v $(PWD)/freezer.db:/data/freezer.db -e DB_PATH=/data/freezer.db $(REGISTRY)/$(BACKEND_IMAGE):$(VERSION)
	@echo "Starting Frontend..."
	docker run -d --rm --name freezo-frontend -p 5000:80 $(REGISTRY)/$(FRONTEND_IMAGE):$(VERSION)
	@echo "App running at http://localhost:5000"

stop-docker:
	docker stop freezo-backend freezo-frontend

reset-db:
	rm -f freezer.db backend/freezer.db
	@echo "Database reset."

test:
	cd backend && go test -v ./...

