.PHONY: run-backend run-frontend

run-backend:
	cd backend && go run main.go

run-frontend:
	cd frontend && npm run dev

test:
	cd backend && go test -v ./...

