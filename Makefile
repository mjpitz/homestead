SKAFFOLD_DEFAULT_REPO ?= ghcr.io/mjpitz
CWD = $(shell pwd)
VERSION ?= latest

dist: .dist
.dist:
	goreleaser --snapshot --skip-publish --rm-dist

docker: .docker
.docker:
	docker build . \
		--tag $(SKAFFOLD_DEFAULT_REPO)/homestead:latest \
		--tag $(SKAFFOLD_DEFAULT_REPO)/homestead:$(VERSION) \
		--file ./deploy/docker/Dockerfile

docker/release:
	docker buildx build . \
		--platform linux/amd64,linux/arm64 \
		--label "org.opencontainers.image.source=https://github.com/mjpitz/homestead" \
		--label "org.opencontainers.image.version=$(VERSION)" \
		--label "org.opencontainers.image.licenses=AGPL-3.0-only" \
		--label "org.opencontainers.image.title=Homestead" \
		--label "org.opencontainers.image.description=" \
		--tag $(SKAFFOLD_DEFAULT_REPO)/homestead:latest \
		--tag $(SKAFFOLD_DEFAULT_REPO)/homestead:$(VERSION) \
		--file ./deploy/docker/Dockerfile \
		--push
