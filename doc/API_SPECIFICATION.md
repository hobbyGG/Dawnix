# Dawnix 工作流引擎 - API 接口说明文档

## 文档概览

本文档为 Dawnix 工作流引擎的前后端接口规范说明。Dawnix 是一个工作流/流程引擎系统，提供HTTP API支持流程定义、流程实例、任务管理等功能。

> 说明：邮件服务节点已插件化，默认关闭。只有在配置中显式开启 `EMAIL_SERVICE_ENABLED=true` 并提供 `SMTP_TOKEN` 时，才允许创建和运行邮件节点。

**API 基础URL**: `http://localhost:8080/api/v1`

---

## 1. 流程定义管理接口 (Definition)

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
        "type": "start_event",
        "name": "开始",
        "candidates": {},
        "properties": null
      },
      {
        "id": "manager_review",
        "type": "user_task",
        "name": "经理审批",
        "candidates": {
          "candidate_users": ["user_id_1", "user_id_2"],
          "candidate_groups": []
        },
        "properties": {
          "assignee_rule": "FIRST_ONE"
        }
      },
      {
        "id": "end",
        "type": "end_event",
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
      "key": "days",
      "type": "number",
      "value": 0
    },
    {
      "key": "reason",
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
| structure.nodes[].type | string | 是 | 节点类型: start_event, end_event, user_task, fork_gateway, join_gateway, xor_gateway, inclusive_gateway；email_service 为可选插件节点，默认关闭 |
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
| form_definition[].key | string | 是 | 表单字段key |
| form_definition[].type | string | 是 | 字段类型: string, number, boolean, etc. |
| form_definition[].value | any | 否 | 字段默认值 |

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
  "total": 25,
  "list": [
    {
      "id": 1,
      "code": "leave_request",
      "version": 1,
      "name": "请假审批流程",
      "structure": {...},
      "form_definition": [...],
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

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
  "id": 1,
  "code": "leave_request",
  "version": 1,
  "name": "请假审批流程",
  "structure": {
    "nodes": [...],
    "edges": [...],
    "viewport": {...}
  },
  "form_definition": [
    {
      "key": "days",
      "type": "number",
      "value": 0
    },
    {
      "key": "reason",
      "type": "string",
      "value": ""
    }
  ],
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**状态码**:
- `200`: 查询成功
- `400`: ID参数错误
- `404`: 流程定义不存在
- `500`: 服务器错误

---

### 1.4 删除流程定义

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
- `404`: 流程定义不存在
- `500`: 服务器错误

---

## 2. 流程实例管理接口 (Instance)

流程实例是基于流程定义创建的具体执行实例，代表一个具体的工作流执行过程。

### 2.1 创建流程实例

**接口**: `POST /api/v1/instance/create`

**功能**: 创建并启动一个新的流程实例

**请求体**:
```json
{
  "process_code": "leave_request",
  "submitter_id": "user_123",
  "form_data": [
    {
      "key": "days",
      "type": "number",
      "value": 3
    },
    {
      "key": "reason",
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
| submitter_id | string | 是 | 发起人ID |
| form_data | array | 否 | 业务表单数据 |
| form_data[].key | string | 是 | 字段key |
| form_data[].type | string | 是 | 字段类型 |
| form_data[].value | any | 是 | 字段值 |
| parent_id | int64 | 否 | 父流程实例ID (子流程场景) |
| parent_node_id | string | 否 | 父流程节点ID (子流程场景) |

**响应体**:
```json
{
  "id": 100
}
```

**状态码**:
- `200`: 实例创建成功
- `400`: 请求参数错误
- `404`: 流程定义不存在
- `500`: 服务器错误

---

### 2.2 获取流程实例列表

**接口**: `GET /api/v1/instance/list`

**功能**: 分页获取流程实例列表

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10，最多100 |

**请求示例**:
```
GET /api/v1/instance/list?page=1&size=20
```

**响应体**:
```json
{
  "total": 50,
  "items": [
    {
      "id": 100,
      "definition_id": 1,
      "process_code": "leave_request",
      "snapshot_structure": {...},
      "parent_id": 0,
      "parent_node_id": "",
      "form_data": [
        {
          "key": "days",
          "type": "number",
          "value": 3
        }
      ],
      "status": "PENDING",
      "submitter_id": "user_123",
      "finished_at": null,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ]
}
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
  "id": 100,
  "definition_id": 1,
  "process_code": "leave_request",
  "snapshot_structure": {
    "nodes": [...],
    "edges": [...]
  },
  "parent_id": 0,
  "parent_node_id": "",
  "form_data": [
    {
      "key": "days",
      "type": "number",
      "value": 3
    },
    {
      "key": "reason",
      "type": "string",
      "value": "年假"
    }
  ],
  "status": "PENDING",
  "submitter_id": "user_123",
  "finished_at": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**状态码**:
- `200`: 查询成功
- `400`: ID参数错误
- `404`: 流程实例不存在
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
- `404`: 流程实例不存在
- `500`: 服务器错误

---

## 3. 任务管理接口 (Task)

任务是在流程实例执行过程中产生的具体工作项，需要分配给用户进行处理。

### 3.1 获取任务详情

**接口**: `GET /api/v1/task/:id`

**功能**: 获取指定任务的详细信息

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
  "id": 200,
  "instance_id": 100,
  "execution_id": 50,
  "node_id": "manager_review",
  "type": "user_task",
  "assignee": "user_456",
  "candidates": ["user_456", "user_789"],
  "status": "PENDING",
  "action": "",
  "comment": "",
  "form_data": [
    {
      "key": "days",
      "type": "number",
      "value": 3
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
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
- `200`: 查询成功
- `400`: ID参数错误
- `404`: 任务不存在
- `500`: 服务器错误

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
      "id": 200,
      "task_name": "经理审批",
      "status": "PENDING",
      "process_title": "请假审批流程",
      "submitter_name": "张三",
      "arrived_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 201,
      "task_name": "部长审批",
      "status": "PENDING",
      "process_title": "请假审批流程",
      "submitter_name": "李四",
      "arrived_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

**状态码**:
- `200`: 查询成功
- `400`: 参数错误
- `500`: 服务器错误

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
      "key": "days",
      "type": "number",
      "value": 3
    },
    {
      "key": "approval_opinion",
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
| form_data[].key | string | 是 | 字段key |
| form_data[].type | string | 是 | 字段类型 |
| form_data[].value | any | 是 | 字段值 |

**响应体**:
```json
{
  "status": "success"
}
```

**状态码**:
- `200`: 任务完成成功，流程继续推进
- `400`: 请求参数错误
- `404`: 任务不存在
- `500`: 服务器错误

---

## 数据结构说明

### FormDataItem (表单数据项)

表单数据采用列表形式，每个项包含key、type和value三个字段。

```json
{
  "key": "field_name",
  "type": "string",
  "value": "field_value"
}
```

支持的字段类型:
- `string`: 字符串
- `number`: 数字
- `boolean`: 布尔值
- `date`: 日期
- `datetime`: 日期时间
- `array`: 数组
- `object`: 对象

### Candidates (候选人)

用户任务的候选人配置

```json
{
  "candidate_users": ["user_id_1", "user_id_2"],
  "candidate_groups": ["group_id_1"]
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

- `start_event` / `end_event`: 流程开始与结束节点
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
- `404`: 资源不存在
- `500`: 服务器内部错误

---

## 前端集成建议

### 1. 基础请求配置

```javascript
const API_BASE = 'http://localhost:8080/api/v1';

const request = async (method, path, data = null) => {
  const url = `${API_BASE}${path}`;
  const options = {
    method,
    headers: { 'Content-Type': 'application/json' },
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

// 创建流程实例
const instance = await request('POST', '/instance/create', {
  process_code: 'leave_request',
  submitter_id: 'user_123',
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

## 版本信息

- API 版本: v1
- 最后更新: 2026-04-19
- 维护者: Dawnix Team
