-include .env

.PHONY: fmt
fmt:
	@go fmt .

.PHONY: lint
lint:
	@golangci-lint run ./...

# REGISTRY_OWNER = docker.io/$(OWNER) where OWNER is defined in .env file
.PHONY: img
img: REGISTRY_OWNER := $(OWNER)
img:
	@docker build -t docker.io/$(REGISTRY_OWNER)/monitor:latest .
	@docker push docker.io/$(REGISTRY_OWNER)/monitor:latest

.PHONY: clean
clean:
	@kind delete cluster

# remove cluster and rebuild image first and the rebuild the cluster and install the helm chart
.PHONY: dev
dev: clean img
	@kind create cluster
	@helm install monitor monitor/ -n monitor --create-namespace