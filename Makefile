.PHONY: test
test:
	gotestsum --format testname

.PHONY: dev
dev:
	air

.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem

.PHONY: tag
tag:
ifndef version
	$(error version is not set. Usage: make tag version=<version_number>)
endif
	git tag v$(version) && git push origin v$(version)

.PHONY: tidy
tidy:
	go mod tidy && go mod vendor && go mod verify && go mod download