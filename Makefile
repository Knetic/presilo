default: build

export GOPATH=$(CURDIR)/
export GOBIN=$(CURDIR)/.temp/

init: clean
	@# Noop, since we don't have any dependencies.
	@:

build: init
	go build .

test:
	go test
	go test -bench=.

clean:
	@rm -rf ./.output/
	@rm -rf ./pkg/

format:
	go fmt . presilo
