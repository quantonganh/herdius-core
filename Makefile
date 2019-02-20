# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

GOPORT=''
GOHOST=''
GOPEERS=''
GOGROUPSIZE=''

GOPARAMETERS=''

ifeq (,$(subst ",,$(PORT)))
    GOPORT=''
else
	GOPARAMETERS :=$(GOPARAMETERS)'-port='$(PORT)
endif


ifeq (,$(subst ",,$(HOST)))
    GOHOST=''
else
	GOPARAMETERS := $(GOPARAMETERS) '-host='$(HOST)
endif


ifeq (,$(subst ",,$(PEERS)))
    GOPEERS=''
else
	GOPARAMETERS := $(GOPARAMETERS) '-peers='$(PEERS)
endif

ifeq (,$(subst ",,$(GROUPSIZE)))
    GOGROUPSIZE=''
else
	GOPARAMETERS := $(GOPARAMETERS) '-groupsize='$(GROUPSIZE)
endif

install:
	$(GOGET) ./...

delete-db-dirs:
	@ rm -R ./herdius/chaindb
	@ rm -R ./herdius/statedb
	@ rm -R ./herdius

create_db_dirs:
	@ mkdir ./herdius && mkdir ./herdius/chaindb/ && mkdir ./herdius/statedb/

build: 
	$(GOBUILD) ./...

run-test: 
	@$(GOTEST) -v ./...

all: install run-test create_db_dirs

start-supervisor: delete-db-dirs create_db_dirs
	@echo "Starting supervisor node"
	@$(GORUN) cmd/herserver/main.go -supervisor=true $(GOPARAMETERS)

start-validator:
	@echo "Starting validator node"
	@$(GORUN) cmd/herserver/main.go$(GOPARAMETERS)
