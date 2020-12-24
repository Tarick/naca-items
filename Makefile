SHELL:=/bin/bash

.DEFAULT_GOAL := help
# put here commands, that have the same name as files in dir
.PHONY: run clean generate build build-images build-and-deploy

BUILD_TAG=$(shell git describe --tags --abbrev=0 HEAD)
BUILD_HASH=$(shell git rev-parse --short HEAD)
BUILD_BRANCH=$(shell git symbolic-ref HEAD |cut -d / -f 3)
BUILD_VERSION=${BUILD_TAG}-${BUILD_HASH}
BUILD_TIME=$(shell date --utc +%F-%H:%m:%SZ)
PACKAGE=naca-items
LDFLAGS=-extldflags=-static -w -s -X ${PACKAGE}/internal/version.Version=${BUILD_VERSION} -X ${PACKAGE}/internal/version.BuildTime=${BUILD_TIME}
# This CONTAINER_REGISTRY must be sourced from environment and it must be FQDN,
# containerd registry plugin doesn't give a shit about short names even if they're present locally, appends docker.io to it
CONTAINER_IMAGE_REGISTRY=${CONTAINER_REGISTRY_FQDN}/items

# ONLY TABS IN THE START OF COMMAND, NO SPACES!
help: ## Return all targets to run
	@echo "build, build-images, deps, build-api, build-worker, build-api-image, build-worker-image, generate-api, build-sql-migrations-image, build-and-deploy, deploy-to-local-k8s"

version: # Return build version
	@echo "${BUILD_VERSION}"

build: deps build-api build-worker # Build all targets

build-images: build-api-image build-worker-image build-sql-migrations-image # Build all docker images

clean: ## Remove build artifacts
	@echo "[INFO] Cleaning build files"
	rm -f build/*

deps:
	@echo "[INFO] Downloading and installing dependencies"
	go mod download

build-api: deps generate
	@echo "[INFO] Building API Server binary"
	go build -ldflags "${LDFLAGS}" -o build/items-api ./cmd/items-api
	@echo "[INFO] Build successful"

generate:
	@echo "[INFO] Running code generations for all packages"
	@go generate ./...
	@echo "[INFO] Finished code generation"

build-api-image:
	@echo "[INFO] Building API container image"
	buildctl build --frontend dockerfile.v0 --opt build-arg:BUILD_VERSION=${BUILD_VERSION} \
	--local context=. --local dockerfile=cmd/items-api  \
	--output type=image,\"name=${CONTAINER_IMAGE_REGISTRY}/items-api:${BUILD_BRANCH}-${BUILD_HASH},${CONTAINER_IMAGE_REGISTRY}/items-api:${BUILD_VERSION}\"
	@echo "[INFO] Image built successfully"

build-worker: deps # Builds worker binary
	@echo "[INFO] Building Worker Server binary"
	go build -ldflags "${LDFLAGS}" -o build/items-worker ./cmd/items-worker
	@echo "[INFO] Build successful"

build-worker-image: # Builds worker container image
	@echo "[INFO] Building Worker container image"
	buildctl build --frontend dockerfile.v0 --opt build-arg:BUILD_VERSION=${BUILD_VERSION} \
	--local context=. --local dockerfile=cmd/items-worker  \
	--output type=image,\"name=${CONTAINER_IMAGE_REGISTRY}/items-worker:${BUILD_BRANCH}-${BUILD_HASH},${CONTAINER_IMAGE_REGISTRY}/items-worker:${BUILD_VERSION}\"
	@echo "[INFO] Image built successfully"

build-sql-migrations-image: # Builds sql-migrations container image
	@echo "[INFO] Building SQL migrations image"
	buildctl build --frontend dockerfile.v0 --opt build-arg:BUILD_VERSION=${BUILD_VERSION} \
	--local context=. --local dockerfile=migrations/  \
	--output type=image,\"name=${CONTAINER_IMAGE_REGISTRY}/items-sql-migrations:${BUILD_BRANCH}-${BUILD_HASH},${CONTAINER_IMAGE_REGISTRY}/items-sql-migrations:${BUILD_VERSION}\"
	@echo "[INFO] Image built successfully"

build-and-deploy: build-images deploy-to-local-k8s ## Builds all images and deploys to local k8s
	@echo "[INFO] built and deployed"

deploy-to-local-k8s: # Deploys built images to local k8s
	@echo "[INFO] Deploying current Items to local k8s service"
	@echo "[INFO] Deleting old SQL migrations"
	helmfile --environment local --selector app_name=items-sql-migrations -f ../naca-ops-config/helm/helmfile.yaml destroy
	@echo "[INFO] Deploying Items images with tag ${BUILD_VERSION}"
	ITEMS_TAG=${BUILD_VERSION} helmfile --environment local --selector tier=naca-items -f ../naca-ops-config/helm/helmfile.yaml sync --skip-deps
	@echo "[INFO] Deployed to local k8s"
