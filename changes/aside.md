# Changes

## Overview

Sibling projects:

- [vue3codegen](#vue3codegen)
- [vue3project](#vue3project)

## Project Structures

### vue3codegen

**Location**: [vue3codegen](../vue3codegen)

```bash
cd vue3codegen && tree --noreport
```

```
.
|-- go.mod
|-- go.sum
|-- main.go
|-- README.md
`-- README.zh.md
```

---

### vue3project

**Location**: [vue3project](../vue3project)

```bash
cd vue3project && tree --noreport
```

```
.
|-- eslint.config.ts
|-- index.html
|-- package-lock.json
|-- package.json
|-- README.md
|-- README.zh.md
|-- src
|   |-- api
|   |   `-- transport.ts
|   |-- App.vue
|   |-- components
|   |   |-- ArticleDemo.vue
|   |   `-- StudentDemo.vue
|   |-- main.ts
|   |-- rpc
|   |   |-- demo1
|   |   |   |-- google
|   |   |   |   |-- api
|   |   |   |   |   |-- field_behavior.ts
|   |   |   |   |   |-- http.ts
|   |   |   |   |   `-- httpbody.ts
|   |   |   |   `-- protobuf
|   |   |   |       |-- any.ts
|   |   |   |       `-- descriptor.ts
|   |   |   `-- student
|   |   |       |-- reason_enum.ts
|   |   |       |-- student.client.ts
|   |   |       `-- student.ts
|   |   `-- demo2
|   |       |-- article
|   |       |   |-- article.client.ts
|   |       |   |-- article.ts
|   |       |   `-- reason_enum.ts
|   |       `-- google
|   |           |-- api
|   |           |   |-- field_behavior.ts
|   |           |   |-- http.ts
|   |           |   `-- httpbody.ts
|   |           `-- protobuf
|   |               |-- any.ts
|   |               `-- descriptor.ts
|   |-- style.css
|   `-- utils
|       |-- error.ts
|       `-- message.ts
|-- tsconfig.app.json
|-- tsconfig.json
|-- tsconfig.node.json
`-- vite.config.ts
```

