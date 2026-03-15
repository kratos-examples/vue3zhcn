# vue3codegen

代码生成工具，从 Kratos 的 proto 文件生成 TypeScript 客户端代码。运行一次，Vue3 前端就能获得调用后端服务所需的全部客户端代码。

---

![grpc-to-http-overview](https://raw.githubusercontent.com/yylego/grpc-to-http/main/assets/grpc-to-http-overview.svg)

---

[ENGLISH README](README.md)

## 解决问题

Kratos 后端在 proto 文件中定义服务（比如 `CreateStudent`、`ListStudents`）。Vue3 前端需要 TypeScript 代码来调用这些服务。手动生成、转换、复制这些文件需要在多个目录之间执行多条命令。这个工具把所有步骤合成一步完成。

## 启动运行

```bash
cd vue3codegen
go run main.go
```

就这样。工具会自动：

1. 找到 `demo1kratos` 和 `demo2kratos` 目录（基于自身的相对路径）
2. 对每个后端：
   - 执行 `make web_api_grpc_ts`，从 proto 文件生成 TypeScript gRPC 客户端
   - 执行 `make web_api_grpc_to_http`，把 gRPC 客户端转换为 HTTP 客户端
   - 把结果复制到 `vue3project/src/rpc/demo1/` 或 `vue3project/src/rpc/demo2/`
   - 执行 `make web_api_cleanup`，清理临时文件

你会看到这样的日志：
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

## 运行时机

当 proto 文件发生变化时运行：
- 新增了服务或方法
- 请求/响应的字段改了
- 删除了 proto 文件

前端的修改（样式、表单、逻辑等）不需要运行。

## 生成产物

运行后，`vue3project/src/rpc/` 会包含：

```
src/rpc/
├── demo1/
│   ├── student/
│   │   ├── student.ts            # 类型：CreateStudentRequest、StudentInfo 等
│   │   ├── student.client.ts     # StudentServiceClient，带类型的方法
│   │   └── reason_enum.ts        # 错误原因枚举
│   └── google/                   # Protobuf 标准类型
└── demo2/
    ├── article/
    │   ├── article.ts            # 类型：CreateArticleRequest、ArticleInfo 等
    │   ├── article.client.ts     # ArticleServiceClient，带类型的方法
    │   └── reason_enum.ts        # 错误原因枚举
    └── google/                   # Protobuf 标准类型
```

这些都是自动生成的文件，不要手动编辑——下次运行会被覆盖。

前端项目的 ESLint 和 Prettier 配置已经把 `src/rpc/` 排除在代码检查和格式化之外。

## 执行流程

```
Proto 文件 (demo1kratos/api/*.proto)
    ↓
make web_api_grpc_ts
    ↓  输出到 demo1kratos/bin/web_api_grpc_ts.out/
    ↓
make web_api_grpc_to_http
    ↓  gRPC 客户端 → HTTP 客户端转换（使用 kratos-vue3）
    ↓
复制到 vue3project/src/rpc/demo1/
    ↓
make web_api_cleanup
    ↓  清理 demo1kratos/bin/ 中的临时文件

（demo2kratos 重复同样的步骤 → vue3project/src/rpc/demo2/）
```

## 前置依赖

`demo1kratos/Makefile` 和 `demo2kratos/Makefile` 必须包含以下目标：

- `web_api_grpc_ts` — 从 proto 生成 TypeScript
- `web_api_grpc_to_http` — gRPC 转 HTTP 客户端
- `web_api_cleanup` — 清理临时文件

工具运行前会检查这些目标是否存在，如果缺少会报错。

## 注意事项

- 这是项目专用工具，硬编码了相对路径，不是通用方案
- 如果你想在自己的项目中使用这种模式，复制 `main.go` 的逻辑并修改路径
- 每次运行会先清理之前的输出再生成新代码

## 相关链接

- [vue3project](../vue3project) — 使用生成代码的前端项目
- [demo1kratos](../demo1kratos) — 学生后端
- [demo2kratos](../demo2kratos) — 文章后端
- [kratos-vue3](https://github.com/yylego/kratos-vue3) — gRPC 转 HTTP 转换库
