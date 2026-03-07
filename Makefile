.PHONY: dev
dev:
	air

.PHONY: tag
tag:
ifndef version
	$(error version is not set. Usage: make tag version=<version_number>)
endif
	git tag v$(version) && git push origin v$(version)

.PHONY: test
test:
	gotestsum --format testname

.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem
