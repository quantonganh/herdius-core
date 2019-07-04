# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

GOPEERS=''
GOPARAMETERS=''

ifeq (,$(subst ,,$(PEERS)))
	GOPEERS=''
else
	GOPARAMETERS := $(GOPARAMETERS) '-peers='$(PEERS)
endif

ifeq (,$(subst ,,$(GROUPSIZE)))
	GOPARAMETERS := $(GOPARAMETERS) '-groupsize=3'
else
	GOPARAMETERS := $(GOPARAMETERS) '-groupsize='$(GROUPSIZE)
endif

ifeq (,$(subst ,,$(PORT)))
	GOPARAMETERS := $(GOPARAMETERS) '-port=0'
else
	GOPARAMETERS := $(GOPARAMETERS) '-port='$(PORT)
endif

ifeq (,$(subst ,,$(WAITTIME)))
	GOPARAMETERS := $(GOPARAMETERS) '-waitTime=15'
else
	GOPARAMETERS := $(GOPARAMETERS) '-waitTime='$(WAITTIME)
endif

ifeq (,$(subst ,,$(ENV)))
	GOPARAMETERS := $(GOPARAMETERS) '-env=dev'
else
	GOPARAMETERS := $(GOPARAMETERS) '-env='$(ENV)
endif

install:
	$(GOGET) ./...

delete-db-dirs:
	@ rm -R ./herdius/chaindb
	@ rm -R ./herdius/statedb
	@ rm -R ./herdius

create_db_dirs:
	@ mkdir -p ./herdius && mkdir -p ./herdius/chaindb/ && mkdir -p ./herdius/statedb/ && mkdir -p ./herdius/syncdb/

build:
	$(GOBUILD) ./...

build-herserver:
	$(GOBUILD) -o ./herserver ./cmd/herserver/main.go

run-test:
	@$(GOTEST) -v ./...

all: install run-test create_db_dirs

start-supervisor: build-herserver
	@echo "Starting supervisor node"$(GOPARAMETERS)
	@./herserver -supervisor=true$(GOPARAMETERS)

start-validator:
	@echo "Starting validator node"
	@./herserver $(GOPARAMETERS)
