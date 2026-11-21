GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

.PHONY: init
# init env
init:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/yylego/kratos-ts-http/cmd/protoc-gen-kratos-ts-http@latest
	npm install -g @protobuf-ts/plugin

.PHONY: config
# generate internal proto
config:
	buf generate --template buf.gen.config.yaml

.PHONY: api
# generate api proto
api:
	buf generate --template buf.gen.yaml

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# generate all
all:
	make api
	make config
	make generate

# 从 proto 一步生成浏览器可直接使用的 TypeScript 客户端，产出即是最终代码，无需二次改写。
# 落到本项目的 bin 下，之后人工拷进前端；`.out` 结尾的目录 Go 的 .gitignore 本就忽略。
# generate web-safe TypeScript clients from proto; the output is the finished code.
# It lands under bin here, ready to be copied into the frontend by hand.
#
# 依赖的两个插件由 `make init` 一并安装。
# Both plugins it needs come from `make init`.
#
# 想一步直达前端，跑 vue3codegen 那条（路径由 Go 算，不必写死在这里）。
# For a straight run into the frontend, use vue3codegen: it works the path out in Go.
web_rpc_ts:
	buf generate --template buf.gen.ts.yaml --include-imports --clean -o ./bin/web_rpc_ts.out

# 删掉本地产物目录本身。平时用不上——生成那条带 `--clean`，每次已是从零开始，
# 而 `.out` 结尾的目录 git 本就忽略；只在想彻底腾空这块磁盘时才用它。
# Removes the local output directory itself. Rarely needed: the generate target passes `--clean`,
# so each run already starts from nothing, and git ignores `.out` dirs anyway.
web_rpc_ts_cleanup:
	rm -rf ./bin/web_rpc_ts.out

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
