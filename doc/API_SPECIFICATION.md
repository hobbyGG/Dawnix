# Dawnix 工作流引擎 API 规范（按当前代码更新）

**API Base URL**: `http://localhost:8080/api/v1`  
**版本**: v1  
**最后更新**: 2026-04-21

## 文档说明

本文件按当前代码实现反向整理，重点补充了“接口 -> 代码位置”的映射，便于快速从 API 跳转到 handler/service/repo。

## 认证与中间件

1. 全局挂载 JWT 中间件：`internal/auth/service/middleware.go`
2. `/api/v1/auth/*` 请求跳过 JWT 校验
3. Workflow 路由（enum/definition/instance/task）还会经过 `InjectUID`：`api/workflow/middleware/uid_middleware.go`

### CORS（当前实现）

- CORS 中间件挂载位置：`cmd/server/manual.go`
- 允许来源读取配置：`server.cors.allow_origins`（见 `configs/dev.yaml`）
- 当前默认允许：`http://localhost:5173`
- 预检请求由 CORS 中间件处理；若 `Origin` 不在允许列表，返回 `403`
- `http://localhost:5173` 与 `http://127.0.0.1:5173` 属于不同源，需分别配置

### 鉴权行为（当前实现）

- 未带 Bearer Token：`401 {"error":"missing bearer token"}`
- Token 无效：`401 {"error":"invalid token"}`
- 上下文缺少 uid（理论上少见）：`401 {"error":"unauthorized"}`

---

## 路由总览（含代码定位）

| 方法 | 路径 | 鉴权 | 入口代码 |
|---|---|---|---|
| POST | `/auth/signup` | 否 | `api/auth/handler.go` |
| POST | `/auth/signin` | 否 | `api/auth/handler.go` |
| POST | `/auth/logout` | 否 | `api/auth/handler.go` |
| GET | `/enum/node-types` | 是 | `api/workflow/enum.go` |
| GET | `/enum/form-types` | 是 | `api/workflow/enum.go` |
| GET | `/enum/approvers` | 是 | `api/workflow/enum.go` |
| POST | `/definition/create` | 是 | `api/workflow/process_definition_handler.go` |
| GET | `/definition/list` | 是 | `api/workflow/process_definition_handler.go` |
| GET | `/definition/:id` | 是 | `api/workflow/process_definition_handler.go` |
| POST | `/definition/update` | 是 | `api/workflow/process_definition_handler.go` |
| POST | `/definition/delete/:id` | 是 | `api/workflow/process_definition_handler.go` |
| POST | `/instance/create` | 是 | `api/workflow/instance_handler.go` |
| GET | `/instance/list` | 是 | `api/workflow/instance_handler.go` |
| GET | `/instance/:id` | 是 | `api/workflow/instance_handler.go` |
| POST | `/instance/delete/:id` | 是 | `api/workflow/instance_handler.go` |
| GET | `/task/:id` | 是 | `api/workflow/task_handler.go` |
| GET | `/task/list` | 是 | `api/workflow/task_handler.go` |
| POST | `/task/complete/:id` | 是 | `api/workflow/task_handler.go` |

> 路由前缀由 `api/router.go` 统一加上 `/api/v1`。

---

## 0. Auth

### 0.1 注册

- **接口**: `POST /api/v1/auth/signup`
- **请求体**:

```json
{
  "username": "admin",
  "password": "password",
  "display_name": "管理员"
}
```

- **响应体**:

```json
{
  "user_id": "1980531045852420096",
  "username": "admin",
  "display_name": "管理员"
}
```

- **状态码**: `200`/`400`/`409`/`500`
- **代码位置**:
  - Handler: `api/auth/handler.go#Signup`
  - Service: `internal/auth/service/service.go#Register`

### 0.2 登录

- **接口**: `POST /api/v1/auth/signin`
- **请求体**:

```json
{
  "username": "admin",
  "password": "password"
}
```

- **响应体**:

```json
{
  "access_token": "<jwt_token>",
  "token_type": "Bearer",
  "expires_at": "2026-04-21T16:00:00+08:00"
}
```

- **状态码**: `200`/`400`/`401`/`500`
- **代码位置**:
  - Handler: `api/auth/handler.go#Signin`
  - Service: `internal/auth/service/service.go#Login`

### 0.3 登出

- **接口**: `POST /api/v1/auth/logout`
- **行为**: 当前为无状态登出，仅返回成功，客户端自行清理 token。
- **响应体**:

```json
{
  "status": "success"
}
```

- **状态码**: `200`
- **代码位置**: `api/auth/handler.go#Logout`

---

## 1. Enum

### 1.1 获取节点类型

- **接口**: `GET /api/v1/enum/node-types`
- **响应**: `{"list":[{"label":"开始节点","value":"start"}, ...]}`
- **说明**: `email_service` 仅在 `biz.features.email_service.enabled=true` 时返回。
- **代码位置**: `api/workflow/enum.go#NodeTypes`

### 1.2 获取表单类型

- **接口**: `GET /api/v1/enum/form-types`
- **响应**: `{"list":[{"label":"单行文本","value":"text_single_line"}, ...]}`
- **代码位置**: `api/workflow/enum.go#FormTypes`

### 1.3 获取审批人枚举（支持模糊搜索）

- **接口**: `GET /api/v1/enum/approvers?keyword=张&limit=20`
- **参数**:
  - `keyword` 可选：按 `display_name` / `user_id` 模糊搜索
  - `limit` 可选：返回条数，范围 `1~100`，默认 `20`
- **响应**:

```json
{
  "list": [
    {
      "label": "张三",
      "value": "1980531045852420096"
    }
  ]
}
```

- **说明**:
  1. 仅返回 `ACTIVE` 用户
  2. `label` 为展示名，`value` 为 `uid`
  3. 推荐在前端审批人选择器中远程搜索并提交 `uid[]`
- **代码位置**: `api/workflow/enum.go#Approvers`

---

## 2. Definition

### 2.1 创建流程定义

- **接口**: `POST /api/v1/definition/create`
- **请求体关键字段**:
  - `name` 必填
  - `structure` 必填（`nodes`、`edges`）
  - `code` 当前代码未做必填校验（建议传）
  - `form_definition` 结构为 `[{id,label,type,value}]`

- **响应体**:

```json
{
  "id": 1
}
```

- **状态码**: `200`/`400`/`500`
- **代码位置**:
  - Handler: `api/workflow/process_definition_handler.go#Create`
  - Service: `internal/workflow/service/process_definition_service.go#CreateProcessDefinition`
  - 校验: `internal/workflow/service/process_definition_service.go#validateAndBuildModel`

### 2.2 流程定义列表

- **接口**: `GET /api/v1/definition/list?page=1&size=10`
- **参数约束**: `page>=1`，`1<=size<=50`
- **响应体**:

```json
{
  "total": 1,
  "list": [
    {
      "ID": 1,
      "Code": "leave_request",
      "Name": "请假审批流程"
    }
  ]
}
```

- **注意**:
  1. `total` 当前是 `len(list)`，不是全量总数。
  2. `page/size` 未设置默认值，建议显式传参。
- **代码位置**:
  - Handler: `api/workflow/process_definition_handler.go#List`
  - Repo: `internal/workflow/data/process_definition_repo.go#List`

### 2.3 流程定义详情

- **接口**: `GET /api/v1/definition/:id`
- **响应**: 直接返回 `domain.ProcessDefinition`（字段为 PascalCase）
- **代码位置**:
  - Handler: `api/workflow/process_definition_handler.go#Detail`
  - Service: `internal/workflow/service/process_definition_service.go#GetProcessDefinitionDetail`

### 2.4 编辑流程定义

- **接口**: `POST /api/v1/definition/update`
- **请求体关键字段**:
  - `id` 必填，流程定义 ID
  - `name` 必填
  - `structure` 必填（`nodes`、`edges`）
- **响应体**:

```json
{
  "status": "updated success"
}
```

- **代码位置**:
  - Handler: `api/workflow/process_definition_handler.go#Update`
  - Service: `internal/workflow/service/process_definition_service.go#UpdateProcessDefinition`

### 2.5 删除流程定义

- **接口**: `POST /api/v1/definition/delete/:id`
- **响应体**:

```json
{
  "status": "deleted success"
}
```

- **代码位置**:
  - Handler: `api/workflow/process_definition_handler.go#Delete`
  - Repo: `internal/workflow/data/process_definition_repo.go#DeleteByID`

---

## 3. Instance

### 3.1 创建流程实例

- **接口**: `POST /api/v1/instance/create`
- **请求体**:

```json
{
  "process_code": "leave_request",
  "form_data": [
    {
      "id": "days",
      "label": "days",
      "type": "number",
      "value": 3
    }
  ],
  "parent_id": 0,
  "parent_node_id": ""
}
```

- **响应体**:

```json
{
  "id": 100
}
```

- **状态码**: `200`/`400`/`500`
- **代码位置**:
  - Handler: `api/workflow/instance_handler.go#Create`
  - Service: `internal/workflow/service/instance_service.go#CreateInstance`
  - Scheduler: `internal/workflow/biz/scheduler.go#StartProcessInstance`

### 3.2 实例列表

- **接口**: `GET /api/v1/instance/list?page=1&size=20`
- **参数约束**: `page>=1`，`1<=size<=100`
- **响应**: 数组，元素为 `domain.ProcessInstance`
- **注意**: `page/size` 未设置默认值，建议显式传参。
- **代码位置**:
  - Handler: `api/workflow/instance_handler.go#List`
  - Repo: `internal/workflow/data/instance_repo.go#List`

### 3.3 实例详情

- **接口**: `GET /api/v1/instance/:id`
- **响应体**:

```json
{
  "inst": {
    "ID": 100,
    "ProcessCode": "leave_request"
  },
  "executions": [
    {
      "ID": 50,
      "NodeID": "manager_review",
      "IsActive": true
    }
  ]
}
```

- **代码位置**:
  - Handler: `api/workflow/instance_handler.go#Detail`
  - Service: `internal/workflow/service/instance_service.go#GetInstanceDetail`

### 3.4 删除实例

- **接口**: `POST /api/v1/instance/delete/:id`
- **响应体**:

```json
{
  "status": "deleted success"
}
```

- **代码位置**:
  - Handler: `api/workflow/instance_handler.go#Delete`
  - Repo: `internal/workflow/data/instance_repo.go#Delete`

---

## 4. Task

### 4.1 任务详情

- **接口**: `GET /api/v1/task/:id`
- **响应**: `domain.TaskDetailView`（以 `process_tasks` 模型字段为主，补充流程标题与提交人）
- **响应体**:

```json
{
  "ID": 200,
  "InstanceID": 100,
  "ExecutionID": 50,
  "NodeID": "manager_review",
  "Type": "user_task",
  "Assignee": "1980531045852420096",
  "Candidates": ["1980531045852420096"],
  "Status": "PENDING",
  "Action": "",
  "Comment": "",
  "FormData": [
    {
      "id": "days",
      "label": "days",
      "type": "number",
      "value": 3
    }
  ],
  "CreatedAt": "2026-04-21T13:00:00+08:00",
  "UpdatedAt": "2026-04-21T13:00:00+08:00",
  "CreatedBy": "",
  "UpdatedBy": "",
  "ProcessTitle": "请假审批流程",
  "SubmitterID": "1980531045852420096"
}
```
- **代码位置**:
  - Handler: `api/workflow/task_handler.go#Detail`
  - Service: `internal/workflow/service/task_service.go#GetTaskDetailView`
  - Repo: `internal/workflow/data/task_repo.go#GetDetailView`

### 4.2 任务列表

- **接口**: `GET /api/v1/task/list?page=1&size=10&scope=my_pending`
- **参数**:
  - `page` 默认 1（handler 内补默认）
  - `size` 默认 10（handler 内补默认，最大 100）
  - `scope` 支持：`my_todo`/`my_pending`/`my_completed`/`all_pending`/`all_completed`

- **响应体**:

```json
{
  "total": 5,
  "tasks": [
    {
      "ID": 200,
      "TaskName": "经理审批",
      "Status": "PENDING",
      "ProcessTitle": "请假审批流程",
      "SubmitterName": "user_123",
      "ArrivedAt": "2026-04-21T13:00:00+08:00"
    }
  ]
}
```

- **当前实现说明**:
  1. `my_todo/my_pending/all_pending` -> 状态过滤为 `PENDING`
  2. `my_completed/all_completed` -> 状态过滤为 `APPROVED`
  3. repo 层始终按当前用户做身份过滤，因此 `all_*` 当前并非“全员可见”
- **代码位置**:
  - Handler: `api/workflow/task_handler.go#List`
  - Service: `internal/workflow/service/task_service.go#ListTasksView`
  - Repo: `internal/workflow/data/task_repo.go#ListWithFilter`

### 4.3 完成任务

- **接口**: `POST /api/v1/task/complete/:id`
- **请求体**:

```json
{
  "action": "agree",
  "comment": "已审批同意",
  "form_data": [
    {
      "id": "approval_opinion",
      "label": "approval_opinion",
      "type": "text_single_line",
      "value": "同意"
    }
  ]
}
```

- **约束**:
  - `action` 必填，且仅支持 `agree/reject`
  - 仅 `PENDING` 任务可完成

- **响应体**:

```json
{
  "status": "success"
}
```

- **代码位置**:
  - Handler: `api/workflow/task_handler.go#Complete`
  - Service: `internal/workflow/service/task_service.go#CompleteTask`
  - Scheduler: `internal/workflow/biz/scheduler.go#CompleteTask`

---

## 数据结构（当前实现）

### FormDataItem

代码定义：`internal/workflow/biz/types.go`

```json
{
  "id": "days",
  "label": "days",
  "type": "number",
  "value": 3
}
```

- `id`: 必填
- `label`: 必填（表达式变量名）
- `type`: 必填（标准值：`text_single_line`/`number`/`single_select`/`date`）
- `value`: 运行时必填；定义时可为空

> 类型别名（如 `string`、`text`、`select`）在当前实现中会被归一化处理，见 `normalizeFormType`。

### ProcessStructure

请求体（Definition Create/Update）里的 `structure.edges` 字段使用 `source/target`。  
落库后领域模型中的边字段为 `source_node/target_node`（`internal/workflow/domain/workflow.go`）。

用户任务节点的审批人配置在 `structure.nodes[].candidates.users`，元素必须是 `uid` 字符串数组。  
建议通过 `GET /api/v1/enum/approvers` 获取前端展示用枚举（`label=姓名`，`value=uid`），提交时仅传 `uid`。

---

## 统一错误格式

大多数错误返回：

```json
{
  "error": "具体错误信息"
}
```

错误写入位置：
- Auth: `api/auth/http_helper.go`
- Workflow: `api/workflow/http_helper.go`
