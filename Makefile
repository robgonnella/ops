PREFIX = $(shell pwd)/build

go_deps = $(shell find . -name '*.go')

.PHONY: all
all: $(PREFIX)/ops

$(PREFIX)/ops: $(go_deps)
	cd cli && go build -ldflags '-s -w' -o $(@)

$(PREFIX)/ops-dev: $(go_deps)
	cd cli && go build -race -ldflags '-s -w' -o $(@)

.PHONY: ops
ops: $(PREFIX)/ops

.PHONY: dev
dev: $(PREFIX)/ops-dev

.PHONY: test
test:
	go test -v -race ./...

.PHONY: mock
mock:
	go generate ./...

.PHONY: clean
clean:
	rm -rf $(PREFIX)
