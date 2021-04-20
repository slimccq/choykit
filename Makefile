# Copyright 2020-present ichenq@outlook.com All rights reserved.
# Distributed under the terms and conditions of the BSD License.
# See accompanying files LICENSE.

GO ?= go

PB_SRC_DIR = protocol
PB_GO_SRC_DIR = pkg/protocol
ALL_PB_SRC = $(wildcard $(PB_SRC_DIR:%=%/*.proto))
ALL_PB_GO_SRC = $(wildcard $(PB_GO_SRC_DIR:%=%/*.pb.go))

THIS_MOD = devpkg.work/choykit
GO_PKG_LIST := $(shell go list ./pkg/...)

export GOBIN = $(shell pwd)/bin

.PHONY: clean build

all: test

$(ALL_PB_GO_SRC): $(ALL_PB_SRC)
	clang-format -i $(ALL_PB_SRC)
	protoc --proto_path=$(PB_SRC_DIR) --gofast_out=$(PB_GO_SRC_DIR)  $(ALL_PB_SRC)

genpb:
	# rm $(ALL_PB_GO_SRC)
	clang-format -i $(ALL_PB_SRC)
	protoc --proto_path=$(PB_SRC_DIR) --gofast_out=$(PB_GO_SRC_DIR) $(ALL_PB_SRC)

test:
	go test -v $(GO_PKG_LIST)

clean:
	go clean
