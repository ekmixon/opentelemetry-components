# All source code and documents, used when checking for misspellings
ALLDOC := $(shell find . \( -name "*.md" -o -name "*.yaml" \) -type f | sort)

GOPATH ?= $(shell go env GOPATH)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

ifeq ($(GOOS), windows)
EXT?=.exe
else
EXT?=
endif

OUTDIR=./build
MODNAME=github.com/observIQ/observiq-collector

LINT=$(GOPATH)/bin/golangci-lint
LINT_TIMEOUT?=5m0s
MISSPELL=$(GOPATH)/bin/misspell

GOBUILDEXTRAENV=GO111MODULE=on CGO_ENABLED=0
GOBUILD=go build
GOINSTALL=go install
GOTEST=go test
GOTOOL=go tool
GOFORMAT=goimports
GOTIDY=go mod tidy

# tool-related commands
.PHONY: install-tools
install-tools:
	$(GOINSTALL) golang.org/x/tools/cmd/goimports
	$(GOINSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.1
	$(GOINSTALL) github.com/client9/misspell/cmd/misspell

# Default build target; making this should build for the current os/arch
.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILDEXTRAENV) \
	$(GOBUILD) $(LDFLAGS) -o $(OUTDIR)/otelcompcol_$(GOOS)_$(GOARCH)$(EXT) ./cmd/otelcompcol

.PHONY: build-all
build-all: amd64_linux amd64_darwin amd64_windows arm_linux

# Other build targets
.PHONY: amd64_linux
amd64_linux:
	GOOS=linux GOARCH=amd64 $(MAKE) otelcompcol

.PHONY: amd64_darwin
amd64_darwin:
	GOOS=darwin GOARCH=amd64 $(MAKE) otelcompcol

.PHONY: arm_linux
arm_linux:
	GOOS=linux GOARCH=arm $(MAKE) otelcompcol

.PHONY: amd64_windows
amd64_windows:
	GOOS=windows GOARCH=amd64 $(MAKE) otelcompcol

.PHONY: lint
lint:
	$(LINT) run --timeout $(LINT_TIMEOUT)

.PHONY: misspell
misspell:
	$(MISSPELL) $(ALLDOC)

.PHONY: misspell-fix
misspell-fix:
	$(MISSPELL) -w $(ALLDOC)

.PHONY: test
test:
	$(GOTEST) -vet off -race ./...

.PHONY: test-with-cover
test-with-cover:
	$(GOTEST) -vet off -cover cover.out ./...
	$(GOTOOL) cover -html=cover.out -o cover.html

.PHONY: check-fmt
check-fmt:
	@GOFMTOUT=`$(GOFORMAT) -d .`; \
		if [ "$$GOFMTOUT" ]; then \
			echo "$(GOFORMAT) SUGGESTED CHANGES:"; \
			echo "$$GOFMTOUT\n"; \
			exit 1; \
		else \
			echo "$(GOFORMAT) completed successfully"; \
		fi

.PHONY: fmt
fmt:
	$(GOFORMAT) -w .

.PHONY: tidy
tidy:
	$(GOTIDY)

# This target performs all checks that CI will do (excluding the build itself)
.PHONY: ci-check
ci-check: check-fmt misspell lint test

.PHONY: clean
clean:
	rm -rf $(OUTDIR)
