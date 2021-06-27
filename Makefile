tag := $(shell git describe --tags)
repo := infrahq/early-access

generate:
	go generate ./...

.PHONY: tools
tools:
	go generate -tags tools tools/tools.go

test:
	go test ./...

.PHONY: docs
docs:
	go run ./internal/docgen

clean:
	rm -rf build release
	docker rm temp

.PHONY: build
build:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o build/infra-Darwin-arm64 -ldflags="-s -w" .
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o build/infra-Darwin-x86_64 -ldflags="-s -w" .

# TODO (jmorganca): find a better way to cross-compile linux & windows binaries
	docker buildx build --load --platform linux/amd64 . -t infrahq/infra:amd64
	docker create --platform linux/amd64 --name temp infrahq/infra:amd64 && docker cp temp:/bin/infra ./build/infra-Linux-amd64 && docker rm -f temp

	docker buildx build --load --platform linux/arm64 . -t infrahq/infra:arm64
	docker create --name temp infrahq/infra:arm64 && docker cp temp:/bin/infra ./build/infra-Linux-arm64 && docker rm -f temp

sign:
	gon .gon.json
	unzip -o -d build build/infra-darwin-binaries.zip
	rm build/infra-darwin-binaries.zip

release:
	make build
	make sign
	-gh release create $(tag) --title $(tag) -n "" -R $(repo)
	gh release upload $(tag) build/* --clobber -R $(repo)

dev/docker:
	kubectl config use-context docker-desktop
	docker build . -t infrahq/infra:dev
	kubectl apply -f ./deploy/dev
	kubectl rollout restart -n infra deployment/infra
	kubectl rollout restart -n infra deployment/infra-engine

release/docker:
	docker buildx build --push --platform linux/amd64,linux/arm64 . -t infrahq/infra
	docker buildx build --push --platform linux/amd64,linux/arm64 . -t infrahq/infra:$(tag:v%=%)
