help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

functions := $(shell find functions -name \*main.go | awk -F'/' '{print $$2}')

check_pre:

build: ## Build golang binaries
	@for function in $(functions) ; do \
		echo "Building $$function"; \
		env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/$$function functions/$$function/*.go ; \
		zip -j bin/$$function.zip bin/$$function; \
	done

deploy: ## Update functions
	@cd terraform; terraform apply;

test:
	@for function in $(functions) ; do \
		go test -v ./functions/$$function/; \
	done