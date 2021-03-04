
GO ?= go

PB_SRC_DIR = protocol
PB_GO_SRC_DIR = pkg/protocol
ALL_PB_SRC = $(wildcard $(PB_SRC_DIR:%=%/*.proto))
ALL_PB_GO_SRC = $(wildcard $(PB_GO_SRC_DIR:%=%/*.pb.go))

GOBIN = $(shell pwd)/bin
GO_MODULE = devpkg.work/choykit
GO_PKG_LIST := $(shell go list ./pkg/...)

.PHONY: clean all genpb

all: build

build: $(ALL_PB_GO_SRC)
	export GOBIN=$(GOBIN)
	go clean
	go install -v $(GO_MODULE)/cmd/choyd

$(ALL_PB_GO_SRC): $(ALL_PB_SRC)
	clang-format -i $(ALL_PB_SRC)
	protoc --proto_path=$(PB_SRC_DIR) --gofast_out=$(PB_GO_SRC_DIR)  $(ALL_PB_SRC)

genpb:
	rm $(ALL_PB_GO_SRC)
	clang-format -i $(ALL_PB_SRC)
	protoc --proto_path=$(PB_SRC_DIR) --gofast_out=$(PB_GO_SRC_DIR) $(ALL_PB_SRC)

test:
	go test -v $(GO_PKG_LIST)
	# $(foreach pkg, $(GO_PKG_LIST), go test -v $(pkg))

clean:
	go clean
