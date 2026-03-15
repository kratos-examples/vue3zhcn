# vue3project

Vue3 + TypeScript + Element Plus 前端项目，连接两个 Kratos 后端服务（StudentService 和 ArticleService），通过 gRPC 转 HTTP 通信。

---

![grpc-to-http-overview](https://raw.githubusercontent.com/yylego/grpc-to-http/main/assets/grpc-to-http-overview.svg)

---

[ENGLISH README](README.md)

## 启动运行

克隆仓库后，需要 3 个终端来运行完整的前后端。

### 终端 1: 启动 StudentService 后端

```bash
cd demo1kratos
go run ./cmd/demo1kratos -conf ./configs
```

你会看到类似输出：
```
msg=[HTTP] server listening on: [::]:8001
msg=[gRPC] server listening on: [::]:9001
```

### 终端 2: 启动 ArticleService 后端

```bash
cd demo2kratos
go run ./cmd/demo2kratos -conf ./configs
```

你会看到：
```
msg=[HTTP] server listening on: [::]:8002
msg=[gRPC] server listening on: [::]:9002
```

两个后端都使用 SQLite 内存数据库，数据只存在内存中——重启服务后数据会全部清空。

### 终端 3: 启动前端

```bash
cd vue3project
npm install
npm run dev
```

你会看到：
```
VITE v8.x.x  ready in xxx ms

➜  Local:   http://localhost:5173/
```

在浏览器中打开 `http://localhost:5173/`。

## 演示效果

页面是浅灰色背景上的一张白色卡片。顶部是标题 "Vue3 + Kratos Demo"，下面有两个标签页：**StudentService** 和 **ArticleService**。

### StudentService 标签页

一个紫色渐变的标题栏显示 "StudentService (demo1kratos :8001)"。

下面是一个白色卡片区域，里面是数据表格。初始状态表格是空的，显示 "No students yet. Click Create to add one."

点击 **Create** → 弹出一个表单对话框：
- Name（必填，失去焦点时验证）
- Age（数字输入框，默认 18）
- Class（选填）

填写字段后点击 **Submit**。成功后，页面顶部出现绿色提示 "Created: id=1, name=xxx"，对话框自动关闭，表格刷新显示新的一行数据。

表格的样式：
- 有边框线（外边框较粗，内部用虚线分隔）
- 斑马纹行（灰白交替）
- 所有单元格内容居中
- 操作按钮：**Select**（查看详情记录到日志）、**Update**（弹出编辑对话框）、**Delete**（删除该行）

点击 **Update** → 弹出相同的表单对话框，预填当前值。编辑后提交。

如果 Name 为空就提交，表单会在字段下方显示红色验证提示 "Name is required"——请求不会被发送。

如果后端返回错误（比如参数无效），会弹出错误对话框，显示：
- 标题带 HTTP 状态码
- Reason（如 "BAD_PARAM"）
- Message（如 "student name is empty"）

底部有一个可折叠的 **Logs** 面板，按时间顺序显示操作历史：
```
[14:30:01] Loaded 3 students
[14:30:05] Created: id=4, name=Alice
[14:30:08] Updated: id=4, name=Bob
[14:30:10] Deleted: id=4
```

### ArticleService 标签页

切换到 ArticleService 标签页。布局和交互方式完全一样，只是字段不同：
- Title（必填）
- Content（多行文本输入框）
- Student ID（选填，关联到某个学生）

表格里有一个 "Student" 列，显示关联的学生 ID，如果没有关联则显示 "-"。

## 实现原理

前端使用 [protobuf-ts](https://github.com/nicolo-ribaudo/protobuf-ts) 生成的 TypeScript 客户端。这些 gRPC 客户端被 [@yylego/grpc-to-http](https://www.npmjs.com/package/@yylego/grpc-to-http) 转换为 HTTP 客户端，浏览器发送标准的 HTTP 请求到 Kratos 后端。

Vite 开发服务器代理请求：
- `/demo1kratos-base/*` → `http://127.0.0.1:8001`（StudentService）
- `/demo2kratos-base/*` → `http://127.0.0.1:8002`（ArticleService）

错误处理：gRPC-to-HTTP 底层使用 axios，所以后端错误都是 AxiosError。`parseError()` 函数提取 Kratos 错误结构（code、reason、message），`showErrorDialog()` 显示错误弹窗。

## 目录结构

```
vue3project/
├── src/
│   ├── api/transport.ts           # 两个 transport 配置（demo1、demo2），设置代理路径
│   ├── components/
│   │   ├── StudentDemo.vue        # 学生增删改查: reactive 表单 + FormRules + ElTable + ElDialog
│   │   └── ArticleDemo.vue        # 文章增删改查: 同样的模式，内容字段用 textarea
│   ├── utils/
│   │   ├── error.ts               # parseError() 解析 Kratos 错误，showErrorDialog() 弹窗显示
│   │   └── message.ts             # showSuccess() 绿色消息提示
│   ├── rpc/                       # 由 vue3codegen 自动生成（不要手动编辑）
│   ├── App.vue                    # ElCard 外框 + ElTabs 标签导航
│   ├── main.ts                    # 注册 Element Plus
│   └── style.css                  # 浅灰色页面背景
├── eslint.config.ts               # ESLint: Vue + TypeScript + Prettier
├── .prettierrc.json               # 无分号、单引号、4 空格缩进、120 字符宽
├── .prettierignore                # 排除 src/rpc/ 不做格式化
└── vite.config.ts                 # API 代理配置
```

## 常用命令

```bash
npm run dev       # 开发服务器（热更新）
npm run build     # 类型检查 + 生产构建
npm run lint      # ESLint 自动修复
npm run format    # Prettier 格式化
```

## 技术栈

Vue 3（Composition API、`<script setup>`）、TypeScript、Element Plus、Vite、protobuf-ts、ESLint + Prettier。

## 适用范围

这是一个演示项目，重点展示前后端集成模式，**不包含**跨服务的数据一致性约束。该演示未处理的场景：

- 创建文章时不校验关联的学生是否存在
- 删除学生时不会级联删除其关联的文章
- 有关联文章的学生也可以被直接删除

在生产系统中，这些数据约束必须严格保障。该项目仅演示 CRUD + gRPC 转 HTTP 的通信模式，不涉及业务逻辑的完整性。

## 相关链接

- [vue3codegen](../vue3codegen) — 从 proto 文件生成 `src/rpc/` 客户端代码
- [demo1kratos](../demo1kratos) — 学生后端（Kratos + GORM + SQLite）
- [demo2kratos](../demo2kratos) — 文章后端（Kratos + GORM + SQLite）
