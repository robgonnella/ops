PREFIX = $(shell pwd)/build

go_deps = $(shell find . -name '*.go')

.PHONY: all
all: $(PREFIX)/opi

$(PREFIX)/opi: $(go_deps)
	cd cli && go build -ldflags '-s -w' -o $(@)

.PHONY: opi
opi: $(PREFIX)/opi

.PHONY: test
test:
	go test -v ./...

.PHONY: mock
mock:
	go generate ./...

.PHONY: clean
clean:
	rm -rf $(PREFIX)
