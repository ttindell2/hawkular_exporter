GO = go
BIN_DIR ?= $(shell pwd)

info:
	@echo "build: Go build"
	@echo "docker: build and run in docker container"
	@echo "gotest: run go tests and reformats"

build_fast:
	$(GO) build -o hawkular_exporter_fast

build:
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -o hawkular_exporter

run: 
	./hawkular_exporter_fast

build_run: buildfast run

docker: build
	sudo docker build -t hawkular_exporter .
