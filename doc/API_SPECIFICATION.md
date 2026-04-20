# Dawnix 工作流引擎 - API 接口说明文档

## 文档概览

本文档为 Dawnix 工作流引擎的前后端接口规范说明。Dawnix 是一个工作流/流程引擎系统，提供HTTP API支持流程定义、流程实例、任务管理等功能。

> 说明：邮件服务节点已插件化，默认关闭。只有在配置中显式开启 `EMAIL_SERVICE_ENABLED=true` 并提供 `SMTP_TOKEN` 时，才允许创建和运行邮件节点。

**API 基础URL**: `http://localhost:8080/api/v1`

**鉴权方式**: Bearer Token（JWT）

> 认证能力遵循 KISS：当前提供注册、登录和登出接口。
> Workflow 相关接口当前直接返回领域对象，响应字段名以 PascalCase 为主（如 `ID`、`CreatedAt`）。

---

## 0. 认证接口 (Auth)

### 0.1 注册 (Signup)

**接口**: `POST /api/v1/auth/signup`

**功能**: 使用本地账号密码注册新用户。

**请求体**:
```json
{
  "username": "admin",
  "password": "password",
  "display_name": "管理员"
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 登录账号（唯一） |
| password | string | 是 | 登录密码 |
| display_name | string | 否 | 显示名称；不传时默认使用 username |

**响应体**:
```json
{
  "user_id": "1980531045852420096",
  "username": "admin",
  "display_name": "管理员"
}
```

**状态码**:
- `200`: 注册成功
- `400`: 请求参数错误
- `409`: 用户名已存在
- `500`: 服务器错误

### 0.2 登录 (Signin)

**接口**: `POST /api/v1/auth/signin`

**功能**: 使用本地账号密码登录，签发访问令牌。

**请求体**:
```json
{
  "username": "admin",
  "password": "password"
}
```

**响应体**:
```json
{
  "access_token": "<jwt_token>",
  "token_type": "Bearer",
  "expires_at": "2026-04-20T12:00:00Z"
}
```

**状态码**:
- `200`: 登录成功
- `400`: 请求参数错误
- `401`: 用户名或密码错误

### 0.3 登出

**接口**: `POST /api/v1/auth/logout`

**功能**: 登出接口（当前为无状态退出，服务端返回成功后由客户端清理 token）。

**响应体**:
```json
{
  "status": "success"
}
```

**状态码**:
- `200`: 登出成功
- `401`: 未登录或 token 无效

---

## 1. 枚举接口 (Enum)

用于给前端提供下拉选项，返回中文展示名和后端英文值。
邮件服务节点是否返回，取决于后端是否开启邮件服务特性。

### 0.1 获取节点类型枚举

**接口**: `GET /api/v1/enum/node-types`

**功能**: 获取当前支持的节点类型列表

**响应体**:
```json
{
  "list": [
    {
      "label": "开始节点",
      "value": "start"
    },
    {
      "label": "结束节点",
      "value": "end"
    },
    {
      "label": "用户任务",
      "value": "user_task"
    },
    {
      "label": "并行分支网关",
      "value": "fork_gateway"
    },
    {
      "label": "并行汇聚网关",
      "value": "join_gateway"
    },
    {
      "label": "排他网关",
      "value": "xor_gateway"
    },
    {
      "label": "包含网关",
      "value": "inclusive_gateway"
    },
    {
      "label": "邮件服务节点",
      "value": "email_service"
    }
  ]
}
```

**字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| list | array | 枚举列表 |
| list[].label | string | 中文展示名，用于前端下拉显示 |
| list[].value | string | 后端英文值，用于提交和保存 |

> 说明：邮件服务节点仅在后端开启邮件服务特性时返回。

**状态码**:
- `200`: 查询成功
- `500`: 服务器错误

---

### 0.2 获取表单类型枚举

**接口**: `GET /api/v1/enum/form-types`

**功能**: 获取当前支持的表单字段类型列表

**响应体**:
```json
{
  "list": [
    {
      "label": "单行文本",
      "value": "text_single_line"
    },
    {
      "label": "数字",
      "value": "number"
    },
    {
      "label": "单选/下拉",
      "value": "single_select"
    },
    {
      "label": "日期",
      "value": "date"
    }
  ]
}
```

**字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| list | array | 枚举列表 |
| list[].label | string | 中文展示名，用于前端下拉显示 |
| list[].value | string | 后端英文值，用于提交和保存 |

**状态码**:
- `200`: 查询成功
- `500`: 服务器错误

---

## 2. 流程定义管理接口 (Definition)

流程定义是工作流的模板，包含流程的结构、节点、连线等信息。

### 1.1 创建流程定义

**接口**: `POST /api/v1/definition/create`

**功能**: 创建一个新的流程定义模板

**请求体**:
```json
{
  "code": "leave_request",
  "name": "请假审批流程",
  "structure": {
    "nodes": [
      {
        "id": "start",
        "type": "start",
        "name": "开始",
        "candidates": {},
        "properties": null
      },
      {
        "id": "manager_review",
        "type": "user_task",
        "name": "经理审批",
        "candidates": {
          "users": ["user_id_1", "user_id_2"]
        },
        "properties": {
          "assignee_rule": "FIRST_ONE"
        }
      },
      {
        "id": "end",
        "type": "end",
        "name": "结束",
        "candidates": {},
        "properties": null
      }
    ],
    "edges": [
      {
        "id": "edge_1",
        "source": "start",
        "target": "manager_review",
        "condition": "",
        "is_default": false
      },
      {
        "id": "edge_2",
        "source": "manager_review",
        "target": "end",
        "condition": "",
        "is_default": false
      }
    ],
    "viewport": {
      "x": 0,
      "y": 0,
      "zoom": 1
    }
  },
  "form_definition": [
    {
      "id": "days",
      "label": "days",
      "type": "number",
      "value": 0
    },
    {
      "id": "reason",
      "label": "reason",
      "type": "string",
      "value": ""
    }
  ]
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 流程代码，用于创建流程实例时的唯一标识 |
| name | string | 是 | 流程名称 |
| structure | object | 是 | 流程结构 (节点和连线) |
| structure.nodes | array | 是 | 节点列表 |
| structure.nodes[].id | string | 是 | 节点唯一ID |
| structure.nodes[].type | string | 是 | 节点类型: start, end, user_task, fork_gateway, join_gateway, xor_gateway, inclusive_gateway；email_service 为可选插件节点，默认关闭 |
| structure.nodes[].name | string | 是 | 节点显示名称 |
| structure.nodes[].candidates | object | 否 | 候选人信息(仅user_task有效) |
| structure.nodes[].properties | object | 否 | 节点特有属性 (JSON格式) |
| structure.edges | array | 是 | 连线列表 |
| structure.edges[].id | string | 是 | 连线ID |
| structure.edges[].source | string | 是 | 源节点ID |
| structure.edges[].target | string | 是 | 目标节点ID |
| structure.edges[].condition | string | 否 | 条件表达式(网关使用) |
| structure.edges[].is_default | boolean | 否 | 是否为默认连线 |
| structure.viewport | object | 否 | 视口状态(用于恢复前端画布状态) |
| form_definition | array | 否 | 表单定义项 |
| form_definition[].id | string | 是 | 表单字段唯一ID |
| form_definition[].label | string | 是 | 表单字段展示名，作为运行时变量名 |
| form_definition[].type | string | 是 | 字段类型：text_single_line, number, single_select, date |
| form_definition[].value | any | 否 | 字段默认值（可选） |

**响应体**:
```json
{
  "id": 1
}
```

**状态码**:
- `200`: 创建成功，返回流程定义ID
- `400`: 请求参数错误
- `500`: 服务器错误

---

### 1.2 获取流程定义列表

**接口**: `GET /api/v1/definition/list`

**功能**: 分页获取流程定义列表

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 是 | 页码，从1开始 |
| size | int | 是 | 每页数量，1-50 |

**请求示例**:
```
GET /api/v1/definition/list?page=1&size=10
```

**响应体**:
```json
{
  "total": 1,
  "list": [
    {
      "ID": 1,
      "CreatedAt": "2024-01-15T10:30:00Z",
      "UpdatedAt": "2024-01-15T10:30:00Z",
      "DeletedAt": null,
      "CreatedBy": "",
      "UpdatedBy": "",
      "Code": "leave_request",
      "Version": 1,
      "Name": "请假审批流程",
      "Structure": {
        "nodes": [
          {
            "id": "start",
            "type": "start",
            "name": "开始"
          }
        ],
        "edges": [
          {
            "id": "edge_1",
            "source_node": "start",
            "target_node": "end",
            "condition": "",
            "is_default": false
          }
        ]
      },
      "FormDefinition": [
        {
          "id": "days",
          "label": "days",
          "type": "number",
          "value": 0
        }
      ],
      "IsActive": true
    }
  ]
}
```

> 说明：`total` 当前实现为本次返回列表长度（`len(list)`），不是全量总记录数。

**状态码**:
- `200`: 查询成功
- `400`: 参数错误
- `500`: 服务器错误

---

### 1.3 获取流程定义详情

**接口**: `GET /api/v1/definition/:id`

**功能**: 获取指定流程定义的详细信息

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | 是 | 流程定义ID |

**请求示例**:
```
GET /api/v1/definition/1
```

**响应体**:
```json
{
  "ID": 1,
  "CreatedAt": "2024-01-15T10:30:00Z",
  "UpdatedAt": "2024-01-15T10:30:00Z",
  "DeletedAt": null,
  "CreatedBy": "",
  "UpdatedBy": "",
  "Code": "leave_request",
  "Version": 1,
  "Name": "请假审批流程",
  "Structure": {
    "nodes": [...],
    "edges": [...],
    "viewport": {...}
  },
  "FormDefinition": [
    {
      "id": "days",
      "label": "days",
      "type": "number",
      "value": 0
    },
    {
      "id": "reason",
      "label": "reason",
      "type": "string",
      "value": ""
    }
  ],
  "IsActive": true
}
```

**状态码**:
- `200`: 查询成功
- `400`: ID参数错误
- `500`: 服务器错误

---

### 1.4 编辑流程定义

**接口**: `PUT /api/v1/definition/:id`

**功能**: 编辑指定流程定义

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | 是 | 流程定义ID |

**请求体**:
```json
{
  "code": "leave_request",
  "name": "请假审批流程(新版)",
  "structure": {
    "nodes": [...],
    "edges": [...],
    "viewport": {"x": 0, "y": 0, "zoom": 1}
  },
  "form_definition": [
    {
      "id": "days",
      "label": "days",
      "type": "number",
      "value": 0
    }
  ]
}
```

**请求示例**:
```
PUT /api/v1/definition/1
```

**响应体**:
```json
{
  "status": "updated success"
}
```

**状态码**:
- `200`: 编辑成功
- `400`: 请求参数错误
- `500`: 服务器错误

---

### 1.5 删除流程定义

**接口**: `DELETE /api/v1/definition/:id`

**功能**: 删除指定的流程定义

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | 是 | 流程定义ID |

**请求示例**:
```
DELETE /api/v1/definition/1
```

**响应体**:
```json
{
  "status": "deleted success"
}
```

**状态码**:
- `200`: 删除成功
- `400`: ID参数错误
- `500`: 服务器错误

---

## 3. 流程实例管理接口 (Instance)

流程实例是基于流程定义创建的具体执行实例，代表一个具体的工作流执行过程。

### 2.1 创建流程实例

**接口**: `POST /api/v1/instance/create`

**功能**: 创建并启动一个新的流程实例

**请求体**:
```json
{
  "process_code": "leave_request",
  "submitter_id": "u_admin",
  "form_data": [
    {
      "id": "days",
      "label": "days",
      "type": "number",
      "value": 3
    },
    {
      "id": "reason",
      "label": "reason",
      "type": "string",
      "value": "年假"
    }
  ],
  "parent_id": 0,
  "parent_node_id": ""
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| process_code | string | 是 | 流程代码，对应ProcessDefinition的code字段 |
| submitter_id | string | 否 | 发起人ID；已登录时以后端 token 注入为准 |
| form_data | array | 否 | 业务表单数据 |
| form_data[].id | string | 是 | 字段唯一ID |
| form_data[].label | string | 是 | 字段展示名，作为运行时变量名 |
| form_data[].type | string | 是 | 字段类型（需与 form_definition 对应字段一致） |
| form_data[].value | any | 是 | 字段值 |
| parent_id | int64 | 否 | 父流程实例ID (子流程场景) |
| parent_node_id | string | 否 | 父流程节点ID (子流程场景) |

> 说明：`form_data` 中不允许提交未在 `form_definition` 声明的字段，命中会返回错误。

**响应体**:
```json
{
  "id": 100
}
```

**状态码**:
- `200`: 实例创建成功
- `400`: 请求参数错误
- `500`: 服务器错误

---

### 2.2 获取流程实例列表

**接口**: `GET /api/v1/instance/list`

**功能**: 分页获取流程实例列表

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| size | int | 否 | 每页数量，最大100 |

**请求示例**:
```
GET /api/v1/instance/list?page=1&size=20
```

**响应体**:
```json
[
  {
    "ID": 100,
    "CreatedAt": "2024-01-15T10:30:00Z",
    "UpdatedAt": "2024-01-15T10:30:00Z",
    "DeletedAt": null,
    "CreatedBy": "",
    "UpdatedBy": "",
    "DefinitionID": 1,
    "ProcessCode": "leave_request",
    "SnapshotStructure": {...},
    "ParentID": 0,
    "ParentNodeID": "",
    "FormData": [...],
    "Status": "PENDING",
    "SubmitterID": "user_123",
    "FinishedAt": null
  }
]
```

**实例状态**:
- `PENDING`: 进行中
- `APPROVED`: 已批准
- `REJECTED`: 已驳回
- `CANCELED`: 已取消
- `SUSPENDED`: 已暂停

**状态码**:
- `200`: 查询成功
- `400`: 参数错误
- `500`: 服务器错误

---

### 2.3 获取流程实例详情

**接口**: `GET /api/v1/instance/:id`

**功能**: 获取指定流程实例的详细信息

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | 是 | 流程实例ID |

**请求示例**:
```
GET /api/v1/instance/100
```

**响应体**:
```json
{
  "inst": {
    "ID": 100,
    "CreatedAt": "2024-01-15T10:30:00Z",
    "UpdatedAt": "2024-01-15T10:30:00Z",
    "DeletedAt": null,
    "CreatedBy": "",
    "UpdatedBy": "",
    "DefinitionID": 1,
    "ProcessCode": "leave_request",
    "SnapshotStructure": {
      "nodes": [...],
      "edges": [...]
    },
    "ParentID": 0,
    "ParentNodeID": "",
    "FormData": [...],
    "Status": "PENDING",
    "SubmitterID": "user_123",
    "FinishedAt": null
  },
  "executions": [
    {
      "ID": 50,
      "CreatedAt": "2024-01-15T10:30:00Z",
      "UpdatedAt": "2024-01-15T10:30:00Z",
      "DeletedAt": null,
      "CreatedBy": "",
      "UpdatedBy": "",
      "InstID": 100,
      "ParentID": 0,
      "NodeID": "manager_review",
      "IsActive": true
    }
  ]
}
```

**状态码**:
- `200`: 查询成功
- `400`: ID参数错误
- `500`: 服务器错误

---

### 2.4 删除流程实例

**接口**: `DELETE /api/v1/instance/:id`

**功能**: 删除指定的流程实例

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | 是 | 流程实例ID |

**请求示例**:
```
DELETE /api/v1/instance/100
```

**响应体**:
```json
{
  "status": "deleted success"
}
```

**状态码**:
- `200`: 删除成功
- `400`: ID参数错误
- `500`: 服务器错误

---

## 4. 任务管理接口 (Task)

任务是在流程实例执行过程中产生的具体工作项，需要分配给用户进行处理。

### 3.1 获取任务详情

**接口**: `GET /api/v1/task/:id`

**功能**: 获取任务视图信息（当前实现返回 TaskView）

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | 是 | 任务ID |

**请求示例**:
```
GET /api/v1/task/200
```

**响应体**:
```json
{
  "ID": 200,
  "TaskName": "经理审批",
  "Status": "PENDING",
  "ProcessTitle": "请假审批流程",
  "SubmitterName": "user_123",
  "ArrivedAt": "2024-01-15T10:30:00Z"
}
```

**任务状态**:
- `PENDING`: 待处理
- `APPROVED`: 已批准
- `REJECTED`: 已驳回
- `TRANSFERRED`: 已转派
- `ROLLED_BACK`: 已撤回
- `CANCELED`: 已取消
- `ABORTED`: 已中止

**任务类型**:
- `user_task`: 用户任务（需要人工审批）
- `service_task`: 服务任务（自动执行；当前文档中的邮件服务能力属于可选插件，不默认启用）
- `receive_task`: 接收任务
- `cc_task`: 抄送任务

**状态码**:
- `200`: 查询成功（业务错误时返回 `{"error":"..."}`）
- `400`: ID参数错误

---

### 3.2 获取任务列表

**接口**: `GET /api/v1/task/list`

**功能**: 分页获取当前用户的任务列表

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10，最多100 |
| scope | string | 否 | 列表范围: my_pending(我的待办), my_completed(我的已办), all_pending(所有待办), all_completed(所有已办) |

**请求示例**:
```
GET /api/v1/task/list?page=1&size=10&scope=my_pending
```

**响应体**:
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
      "ArrivedAt": "2024-01-15T10:30:00Z"
    },
    {
      "ID": 201,
      "TaskName": "部长审批",
      "Status": "PENDING",
      "ProcessTitle": "请假审批流程",
      "SubmitterName": "user_456",
      "ArrivedAt": "2024-01-15T11:00:00Z"
    }
  ]
}
```

> 说明：`scope` 除 `my_pending/my_completed/all_pending/all_completed` 外，还支持 `my_todo`（默认值）。

**状态码**:
- `200`: 查询成功（业务错误时返回 `{"error":"..."}`）
- `400`: 参数错误
- `401`: 未认证或 token 无效

---

### 3.3 完成任务

**接口**: `POST /api/v1/task/complete/:id`

**功能**: 完成指定的任务(审批/驳回等)

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | 是 | 任务ID |

**请求体**:
```json
{
  "action": "agree",
  "comment": "已审批同意",
  "form_data": [
    {
      "id": "days",
      "label": "days",
      "type": "number",
      "value": 3
    },
    {
      "id": "approval_opinion",
      "label": "approval_opinion",
      "type": "string",
      "value": "同意请假"
    }
  ]
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 操作类型: agree(同意), reject(驳回) |
| comment | string | 否 | 审批意见/备注 |
| form_data | array | 否 | 提交的表单数据 |
| form_data[].id | string | 是 | 字段唯一ID |
| form_data[].label | string | 是 | 字段展示名，作为运行时变量名 |
| form_data[].type | string | 是 | 字段类型（需与 form_definition 对应字段一致） |
| form_data[].value | any | 是 | 字段值 |

> 说明：`form_data` 中不允许提交未在 `form_definition` 声明的字段，命中会返回错误。

**响应体**:
```json
{
  "status": "success"
}
```

**状态码**:
- `200`: 任务完成成功（业务错误时返回 `{"error":"..."}`）
- `400`: 请求参数错误
- `401`: 未认证或 token 无效

---

## 数据结构说明

### FormDataItem (表单数据项)

表单数据采用列表形式，每个项包含id、label、type和value四个字段。

- `id`：字段的唯一标识，用于合并和定位同一项
- `label`：字段展示名，同时作为运行时表达式变量名

```json
{
  "id": "field_001",
  "label": "field_name",
  "type": "string",
  "value": "field_value"
}
```

支持的字段类型（当前版本）:
- `text_single_line`: 单行文本
- `number`: 数字
- `single_select`: 单选/下拉
- `date`: 日期（RFC3339 字符串）

当前版本暂不支持：联系人（成员选择器）、附件、图片等复杂类型。

### Candidates (候选人)

用户任务的候选人配置

```json
{
  "users": ["user_id_1", "user_id_2"]
}
```

### ProcessStructure (流程结构)

包含节点和连线的完整流程定义

```json
{
  "nodes": [...],
  "edges": [...],
  "viewport": {
    "x": 0,
    "y": 0,
    "zoom": 1
  }
}
```

### 节点能力说明

- `start` / `end`: 流程开始与结束节点
- `user_task`: 用户任务节点，需要人工审批
- `fork_gateway` / `join_gateway` / `xor_gateway` / `inclusive_gateway`: 网关节点
- `email_service`: 邮件服务节点，默认关闭；启用后会投递邮件任务到 Redis，由 worker 消费发送

如果需要使用邮件服务节点，请在运行配置中开启 `EMAIL_SERVICE_ENABLED=true`，并同步配置 `SMTP_TOKEN`、`SMTP_EMAIL`、`REDIS_ADDR`。

---

## 错误处理

所有接口的错误响应格式统一为:

```json
{
  "error": "错误信息说明"
}
```

常见HTTP状态码:
- `200`: 请求成功
- `400`: 请求参数错误或验证失败
- `401`: 未认证或令牌无效
- `404`: 资源不存在
- `500`: 服务器内部错误

---

## 身份字段迁移说明

为保持 Auth 与 Workflow 一致性，系统内用户身份字段统一为 `string`。

### 统一约束

- `submitter_id` 使用 string
- 任务 `assignee` 使用 string
- 任务 `candidates` 数组成员使用 string
- JWT `sub` 与内部 `uid` 使用 string
- `user_id` 作为 Auth 主身份键，使用 string

### 兼容策略

- 历史整型身份值在迁移时转为字符串存储
- 业务层不再接收 `int64` 用户身份字段
- 新接口与中间件仅向下游传递 string 类型身份

### 迁移建议顺序

1. 先做数据库列类型与数据转换
2. 再切换应用层 DTO 和领域模型
3. 最后清理遗留整型身份字段

---

## 前端集成建议

### 1. 基础请求配置

```javascript
const API_BASE = 'http://localhost:8080/api/v1';

const request = async (method, path, data = null) => {
  const url = `${API_BASE}${path}`;
  const token = localStorage.getItem('access_token');
  const options = {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  };
  
  if (data) {
    options.body = JSON.stringify(data);
  }
  
  const res = await fetch(url, options);
  return res.json();
};
```

### 2. 常用API调用示例

```javascript
// 获取流程定义列表
const definitions = await request('GET', '/definition/list?page=1&size=10');

// 登录（成功后保存 access_token）
const loginResp = await request('POST', '/auth/signin', {
  username: 'admin',
  password: 'password'
});
localStorage.setItem('access_token', loginResp.access_token);

// 创建流程实例
const instance = await request('POST', '/instance/create', {
  process_code: 'leave_request',
  form_data: [...]
});

// 获取待办任务列表
const tasks = await request('GET', '/task/list?scope=my_pending&page=1&size=10');

// 完成任务
await request('POST', `/task/complete/${taskId}`, {
  action: 'agree',
  comment: '已批准',
  form_data: [...]
});
```

### 3. 流程交互流程

1. **发起流程**: 调用 `POST /instance/create` 创建实例
2. **查看待办**: 调用 `GET /task/list` 获取当前用户待办任务
3. **查看任务详情**: 调用 `GET /task/:id` 查看具体任务
4. **处理任务**: 调用 `POST /task/complete/:id` 完成任务
5. **查看流程详情**: 调用 `GET /instance/:id` 查看流程进展

---

## 联调清单 (KISS)

### 1. 登录获取 Token

1. 调用 `POST /api/v1/auth/signin`
2. 保存 `access_token`
3. 后续请求统一携带 `Authorization: Bearer <token>`

### 2. 访问受保护接口

1. 使用 token 调用 `GET /api/v1/task/list`
2. 使用 token 调用 `POST /api/v1/instance/create`
3. 校验后端以 token 中用户身份写入 `submitter_id`

### 3. 未登录校验

1. 不带 token 调用 `GET /api/v1/task/list`
2. 预期返回 `401`

### 4. 登出

1. 调用 `POST /api/v1/auth/logout`
2. 客户端清理本地 token
3. 清理后再次访问受保护接口应返回 `401`

---

## 版本信息

- API 版本: v1
- 最后更新: 2026-04-20
- 维护者: Dawnix Team
