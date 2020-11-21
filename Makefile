SHELL:=/bin/bash

.DEFAULT_GOAL := help
# put here commands, that have the same name as files in dir
.PHONY: run clean generate build docker_build docker_push

BUILD_TAG=$(shell git describe --tags --abbrev=0 HEAD)
BUILD_HASH=$(shell git rev-parse --short HEAD)
BUILD_BRANCH=$(shell git symbolic-ref HEAD |cut -d / -f 3)
BUILD_VERSION="${BUILD_TAG}-${BUILD_HASH}"
BUILD_TIME=$(shell date --utc +%F-%H:%m:%SZ)
PACKAGE=naca-items
LDFLAGS=-extldflags=-static -w -s -X ${PACKAGE}/internal/version.Version=${BUILD_VERSION} -X ${PACKAGE}/internal/version.BuildTime=${BUILD_TIME}
CONTAINER_IMAGE_REGISTRY=local/items

help:
	@echo "build, build-images, deps, build-api, build-worker, build-api-image, build-worker-image, generate-api, build-sql-migrations-image, deploy-to-local-k8s"

version:
	@echo "${BUILD_VERSION}"

# ONLY TABS IN THE START OF COMMAND, NO SPACES!
build: deps build-api build-worker

build-images: build-api-image build-worker-image build-sql-migrations-image

clean:
	@echo "[INFO] Cleaning build files"
	rm -f build/*

deps:
	@echo "[INFO] Downloading and installing dependencies"
	go mod download

build-api: deps generate-api
	@echo "[INFO] Building API Server binary"
	go build -ldflags "${LDFLAGS}" -o build/items-api ./cmd/items-api
	@echo "[INFO] Build successful"

generate-api:
	@echo "[INFO] Running code generations for API"
	go generate cmd/items-api/main.go

build-api-image:
	@echo "[INFO] Building API container image"
	# docker build -t ${CONTAINER_IMAGE_REGISTRY}/items-api:${BUILD_BRANCH}-${BUILD_HASH} \
	# -t ${CONTAINER_IMAGE_REGISTRY}/items-api:${BUILD_VERSION} \
	# --build-arg BUILD_VERSION=${BUILD_VERSION} -f cmd/items-api/Dockerfile .

build-worker: deps
	@echo "[INFO] Building Worker Server binary"
	go build -ldflags "${LDFLAGS}" -o build/items-worker ./cmd/items-worker
	@echo "[INFO] Build successful"

build-worker-image:
	@echo "[INFO] Building Worker container image"
	docker build -t ${CONTAINER_IMAGE_REGISTRY}/items-worker:${BUILD_BRANCH}-${BUILD_HASH} \
	-t ${CONTAINER_IMAGE_REGISTRY}/items-worker:${BUILD_VERSION} \
	--build-arg BUILD_VERSION=${BUILD_VERSION} -f cmd/items-worker/Dockerfile .

build-sql-migrations-image:
	@echo "[INFO] Building SQL migrations image"
	docker build -t ${CONTAINER_IMAGE_REGISTRY}/items-sql-migrations:${BUILD_BRANCH}-${BUILD_HASH} \
	-t ${CONTAINER_IMAGE_REGISTRY}/items-sql-migrations:${BUILD_VERSION} \
	-f migrations/Dockerfile .

deploy-to-local-k8s: build-images
	@echo "[INFO] Deploying current Items to local k8s service"
	@echo "[INFO] Deleting old SQL migrations"
	helmfile --environment local --selector app_name=items-sql-migrations -f ../naca-ops-config/helm/helmfile.yaml destroy
	@echo "[INFO] Deploying Items images with tag ${BUILD_VERSION}"
	ITEMS_TAG=${BUILD_VERSION} helmfile --environment local --selector tier=naca-items -f ../naca-ops-config/helm/helmfile.yaml sync --skip-deps
