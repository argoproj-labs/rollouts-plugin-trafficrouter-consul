.PHONY: go_lint
go_lint:
	golangci-lint run ./...

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(TARGETARCH) go build -v -o rollouts-plugin-trafficrouter-consul ./

docker:
	docker build . -t rollouts-plugin-trafficrouter-consul
	docker tag rollouts-plugin-trafficrouter-consul wilko1989/rollouts-plugin-trafficrouter-consul:latest
	docker push wilko1989/rollouts-plugin-trafficrouter-consul:latest
