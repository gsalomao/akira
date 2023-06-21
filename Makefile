# Copyright 2023 Gustavo Salomao
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Project parameters
BUILD_DIR       = build
BIN_DIR			= ${BUILD_DIR}/bin
COVERAGE_DIR    = ${BUILD_DIR}/coverage

# Colors
GREEN       := $(shell tput -Txterm setaf 2)
YELLOW      := $(shell tput -Txterm setaf 3)
WHITE       := $(shell tput -Txterm setaf 7)
CYAN        := $(shell tput -Txterm setaf 6)
RESET       := \033[0m
BOLD        := \033[0;1m

.PHONY: all
all: help

## Setup
.PHONY: init
init: ## Initialize project
	$(call print_task,"Installing Git hooks")
	@cp scripts/githooks/* .git/hooks
	@chmod +x .git/hooks/*
	$(call print_task_result,"Installing Git hooks","done")

## Build
.PHONY: build
build: ## Build application
	$(call print_task,"Building application")
	@mkdir -p ${BIN_DIR}
	@for dir in `ls examples`; do \
		echo "    Building example '$$dir'..."; \
		go build -o ${BIN_DIR}/$$dir examples/$$dir/main.go; \
	done
	$(call print_task_result,"Building application","done")

.PHONE: update
update: ## Update dependencies
	$(call print_task,"Updating dependencies")
	@go get -u ./...
	@go mod tidy
	$(call print_task_result,"Updating dependencies","done")

.PHONY: clean
clean: ## Clean build folder
	$(call print_task,"Cleaning build folder")
	@go clean
	@rm -rf ${BUILD_DIR}
	$(call print_task_result,"Cleaning build folder","done")

## Run
.PHONY: start
start: ## Start basic example
	$(call print_task,"Starting basic example")
	@$(BIN_DIR)/basic

.PHONY: start-profile
start-profile: ## Start basic example with CPU/Memory profiler
	$(call print_task,"Starting basic example in profiling mode")
	@$(BIN_DIR)/basic --profile

## Test
.PHONY: unit
unit: ## Run unit tests
	$(call print_task,"Running unit tests")
	@gotestsum --format testname --packages ./... -- -timeout 10s -race
	$(call print_task_result,"Running unit tests","done")

.PHONY: unit-dev
unit-dev: ## Run unit tests in development mode
	$(call print_task,"Running unit tests in development mode")
	@gotestsum --format testname --packages ./... --watch -- -timeout 5s -race

.PHONY: coverage
coverage: ## Run unit tests with coverage report
	$(call print_task,"Running unit tests")
	@rm -rf ${COVERAGE_DIR}
	@mkdir -p ${COVERAGE_DIR}
	@go test -timeout 10s -cover -covermode=atomic -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(call print_task_result,"Running unit tests","done")

	$(call print_task,"Generating coverage report")
	@go tool cover -func $(COVERAGE_DIR)/coverage.out
	$(call print_task_result,"Generating coverage report","done")

.PHONY: coverage-html
coverage-html: coverage ## Open the coverage report in the browser
	$(call print_task,"Opening coverage report")
	@go tool cover -html $(COVERAGE_DIR)/coverage.out

## Analyze
.PHONY: inspect
inspect: vet lint ## Run vet and lint commands

.PHONY: vet
vet: ## Examine source code
	$(call print_task,"Examining source code")
	@go vet ./...
	$(call print_task_result,"Examining source code","done")

.PHONY: fmt
fmt: ## Format source code
	$(call print_task,"Formatting source code")
	@go fmt ./...
	$(call print_task_result,"Formatting source code","done")

.PHONY: lint
lint: ## Lint source code
	$(call print_task,"Linting source code")
	@golint  -set_exit_status $(go list ./...)
	@golangci-lint run $(go list ./...)
	$(call print_task_result,"Linting source code","done")

.PHONY: complexity
complexity: ## Calculates cyclomatic complexity
	$(call print_task,"Calculating cyclomatic complexity")
	@gocyclo -top 10 .
	@gocyclo -over 12 -avg .
	$(call print_task_result,"Calculating cyclomatic complexity","done")

## Help
.PHONY: help
help: ## Show this help
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z0-9_-]+:.*?##.*$$/) { \
			printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
			else if (/^## .*$$/) { \
				printf "  ${CYAN}%s:${RESET}\n", substr($$1,4)\
			} \
		}' $(MAKEFILE_LIST)

define print_task
	@printf "${CYAN}==>${BOLD} %s...${RESET}\n" $(1)
endef

define print_task_result
	@printf "${CYAN}==> %s... %s${RESET}\n" $(1) $(2)
endef
