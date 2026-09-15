PROJECT_NAME := "ftv-portal"

.PHONY: app docker docker-run

app: ## make production build
	@npm run build

run: ## run development server
	@npm run dev

docker: ## build docker image
	@docker build -f Dockerfile -t $(PROJECT_NAME):latest-dev ../..

docker-run: ## run docker container
	@docker run -p 8080:8080 $(PROJECT_NAME):latest-dev