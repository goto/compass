.PHONY: all build test clean install proto
all: build

# HELP sourced from https://gist.github.com/prwhite/8168133

# Add help text after each target name starting with '\#\#'
# A category can be added with @category

HELP_FUNC = \
    %help; \
    while(<>) { \
        if(/^([a-z0-9_-]+):.*\#\#(?:@(\w+))?\s(.*)$$/) { \
            push(@{$$help{$$2}}, [$$1, $$3]); \
        } \
    }; \
    print "usage: make [target]\n\n"; \
    for ( sort keys %help ) { \
        print "$$_:\n"; \
        printf("  %-30s %s\n", $$_->[0], $$_->[1]) for @{$$help{$$_}}; \
        print "\n"; \
    }

help:           ##@help show this help
	@perl -e '$(HELP_FUNC)' $(MAKEFILE_LIST)

NAME="github.com/goto/compass"
VERSION=$(shell git describe --always --tags 2>/dev/null)
COVERFILE="/tmp/compass.coverprofile"
PROTON_COMMIT := "a6b2821e8ddd1127a63d3b376f860990d58931da"

TOOLS_MOD_DIR = ./tools
TOOLS_DIR = $(abspath ./.tools)

$(TOOLS_DIR)/buf: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/buf github.com/bufbuild/buf/cmd/buf

$(TOOLS_DIR)/golangci-lint: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

$(TOOLS_DIR)/mockery: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/mockery github.com/vektra/mockery/v2

$(TOOLS_DIR)/gofumpt: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/gofumpt mvdan.cc/gofumpt

$(TOOLS_DIR)/gci: $(TOOLS_MOD_DIR)/go.mod $(TOOLS_MOD_DIR)/go.sum $(TOOLS_MOD_DIR)/tools.go
	cd $(TOOLS_MOD_DIR) && \
	go build -o $(TOOLS_DIR)/gci github.com/daixiang0/gci

# DEV SETUP #############

imports: $(TOOLS_DIR)/gci ##@dev_setup does a goimports
	$(TOOLS_DIR)/gci write ./ --section standard --section default --skip-generated

fmt: $(TOOLS_DIR)/gofumpt imports ##@dev_setup does a go fmt (stricter variant)
	$(TOOLS_DIR)/gofumpt -l -w -extra .

lint: $(TOOLS_DIR)/golangci-lint ##@dev_setup lint source
	$(TOOLS_DIR)/golangci-lint --config=".golangci-prod.toml" --new-from-rev=HEAD~1 --max-same-issues=0 --max-issues-per-linter=0 run

# BUILD #############

proto: $(TOOLS_DIR)/buf ## Generate the protobuf files
	@echo " > generating protobuf from goto/proton"
	$(TOOLS_DIR)/buf generate https://github.com/goto/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path gotocompany/compass -v
	@echo " > protobuf compilation finished"

generate: $(TOOLS_DIR)/mockery ## Run all go generate in the code base
	go generate ./...

clean:  ## Clean the build artifacts
	rm -rf compass dist/

build: ## Build the compass binary
	go build -ldflags "-X ${NAME}/cli.Version=${VERSION}"

# TESTS #############

test: ## Run the tests
	go test -race ./... -coverprofile=coverage.txt

test-coverage: test ## Print the code coverage
	go tool cover -html=coverage.txt -o cover.html

e2e-test: ## Run all e2e tests
	go test ./test/... --tags=e2e

# DOCS #############

clean-doc:
	@echo "> cleaning up auto-generated docs"
	@rm -rf ./docs/docs/reference/cli.md
	@rm -f ./docs/docs/reference/api.md

update-swagger-md:
	npx swagger-markdown -i proto/compass.swagger.yaml -o docs/docs/reference/api.md

doc: clean-doc update-swagger-md ## Generate api and cli references
	@echo "> generate cli docs"
	@go run . reference --plain | sed '1 s,.*,# CLI,' > ./docs/docs/reference/cli.md

