PREFIX = $(shell pwd)/build

go_deps = $(shell find . -name '*.go')

.PHONY: all
all: $(PREFIX)/ops

$(PREFIX)/ops: $(go_deps)
	cd cli && go build -ldflags '-s -w' -o $(@)

.PHONY: ops
ops: $(PREFIX)/ops

.PHONY: test
test:
	go test -v ./...

.PHONY: mock
mock:
	go generate ./...

.PHONY: clean
clean:
	rm -rf $(PREFIX)
