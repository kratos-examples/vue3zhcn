# vue3codegen

Code generation command that produces TypeScript client code from Kratos proto files. Run it once, and the Vue3 frontend gets complete client code to invoke backend services.

---

![grpc-to-http-overview](https://raw.githubusercontent.com/yylego/grpc-to-http/main/assets/grpc-to-http-overview.svg)

---

[中文说明](README.zh.md)

## Motivation

The Kratos backends define services in proto files (like `CreateStudent`, `ListStudents`). The Vue3 frontend needs TypeScript code to invoke these services. Generating, converting, and copying the files takes multiple commands across multiple directories. This command does everything in one step.

## Usage

```bash
cd vue3codegen
go run main.go
```

That's it. The command:

1. Locates `demo1kratos` and `demo2kratos` directories (paths relative to itself)
2. Each backend:
   - Runs `make web_api_grpc_ts` to generate TypeScript gRPC clients from proto files
   - Runs `make web_api_grpc_to_http` to convert gRPC clients to HTTP clients
   - Copies the results into `vue3project/src/rpc/demo1/` and `vue3project/src/rpc/demo2/`
   - Runs `make web_api_cleanup` to remove temp files

Output looks like:
```
=== Vue3 Client Code Gen Workflow Start ===
Backend project: .../demo1kratos
Makefile targets verified
Generating TypeScript gRPC clients...
TypeScript gRPC clients generated
Converting gRPC clients to HTTP clients...
Conversion completed
Syncing converted files...
File sync completed
Cleaning up temp files...
Cleanup completed
Backend project: .../demo2kratos
...
=== WORKFLOW FINISHED SUCCESS! ===
```

## Run Timing

Run this command when proto files change:
- New service / method added
- Request/response fields changed
- Proto file deleted

You do NOT need to run it when frontend code changes (styling, forms, logic, etc.).

## Output

Once done, `vue3project/src/rpc/` contains:

```
src/rpc/
├── demo1/
│   ├── student/
│   │   ├── student.ts            # Types: CreateStudentRequest, StudentInfo, etc.
│   │   ├── student.client.ts     # StudentServiceClient with typed methods
│   │   └── reason_enum.ts        # Reason codes
│   └── google/                   # Protobuf standard types
└── demo2/
    ├── article/
    │   ├── article.ts            # Types: CreateArticleRequest, ArticleInfo, etc.
    │   ├── article.client.ts     # ArticleServiceClient with typed methods
    │   └── reason_enum.ts        # Reason codes
    └── google/                   # Protobuf standard types
```

These are auto-generated files. Do not edit them — next execution overwrites them.

The frontend's ESLint and Prettier configs exclude `src/rpc/` from linting and formatting.

## The Pipeline

```
Proto files (demo1kratos/api/*.proto)
    ↓
make web_api_grpc_ts
    ↓  Outputs to demo1kratos/bin/web_api_grpc_ts.out/
    ↓
make web_api_grpc_to_http
    ↓  Converts gRPC clients → HTTP clients (using kratos-vue3)
    ↓
Copy to vue3project/src/rpc/demo1/
    ↓
make web_api_cleanup
    ↓  Removes temp files from demo1kratos/bin/

(Same steps repeat: demo2kratos → vue3project/src/rpc/demo2/)
```

## Prerequisites

The `demo1kratos/Makefile` and `demo2kratos/Makefile` must have these targets:

- `web_api_grpc_ts` — generates TypeScript from proto
- `web_api_grpc_to_http` — converts gRPC to HTTP clients
- `web_api_cleanup` — cleans temp files

The command checks these targets before execution. If a target is missing, it fails with an explicit message.

## Notes

- This is a project-specific command with hardcoded paths, not a generic solution
- To use this pattern in a different project, duplicate and adapt the logic in `main.go`
- Each execution cleans previous output before generating new code

## See Also

- [vue3project](../vue3project) — The frontend that uses the generated code
- [demo1kratos](../demo1kratos) — Student backend
- [demo2kratos](../demo2kratos) — Article backend
- [kratos-vue3](https://github.com/yylego/kratos-vue3) — The gRPC-to-HTTP conversion package
