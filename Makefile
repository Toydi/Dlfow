GOCMD=go
GOBUILD=GOOS=linux GOARCH=amd64 $(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get


OUTPUT_DIR=build/_output

MAIN_FILE=cmd/manager/main.go
OPERATOR_NAME=df-operator

DOCKER_IMAGE=registry.njuics.cn/qr/dlflow	
.PHONY: build
all: build

build: build-operator

build-operator:
	$(GOBUILD) -o $(OUTPUT_DIR)/bin/$(OPERATOR_NAME) -v $(MAIN_FILE)

.PHONY: test
test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -rf $(OUTPUT_DIR)/bin/$(OPERATOR_NAME)

deps:
	dep ensure -v

install:
	cp $(OPERATOR_NAME) /usr/local/bin

image:
	mkdir -p $(OUTPUT_DIR)
	docker build -t $(DOCKER_IMAGE) -f build/Dockerfile .

ui:
	mkdir -p $(OUTPUT_DIR)
