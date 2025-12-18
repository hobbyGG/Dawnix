
## 项目结构说明
```
/flow-engine
├── /cmd
│   └── /server
│       ├── main.go             # 入口
│       └── wire.go             # Wire 定义
│
├── /configs                    # 配置文件
│
├── /internal
│   │
│   ├── /api                    # [接入层] HTTP Handler (Controller)
│   │   ├── request.go          # 请求参数结构体 (DTO)
│   │   └── process_handler.go  # func Create(c *gin.Context)
│   │
│   ├── /service                # [业务层] Business Logic (Use Case)
│   │   │                       # ⚠️ 这里直接用 Struct，不搞 interface
│   │   └── process_service.go  # type ProcessService struct
│   │
│   ├── /biz                    # [领域层] 核心定义 (Entities & Interfaces)
│   │   ├── /model              # 数据库模型 (GORM struct)
│   │   └── /port               # ⚠️ 接口定义在这里！(Repo, MQ, AI 接口)
│   │       ├── repository.go   # type ProcessRepo interface
│   │       └── component.go    # type EventBus interface, AIGateway interface
│   │
│   ├── /data                   # [数据层] 实现 biz/port 中的 Repo 接口
│   │   └── process_repo.go     # 基于 GORM 的实现
│   │
│   ├── /component              # [组件层] 实现 biz/port 中的 Component 接口 (Infra)
│   │   ├── /mq                 # RabbitMQ / Memory 实现
│   │   └── /ai                 # OpenAI / NoOp 实现
│   │
│   └── /app                    # [组装层] Wire/Manual DI
│       └── app.go              # 组装 Service, Data, Component
│
└── /pkg                        # [公共库] 
    └── /xerr                   # 错误码
```

## MVP版本功能说明

#### 第一阶段：静态定义 (Definition) —— "把图纸存下来"

**目标**：打通 `ProcessDefinition` 的 CRUD，让前端设计器有后端接口可用。

1. **创建流程 (Create)**
	- **后端**：实现 `CreateProcessDefinition` 接口。
	- **逻辑**：接收前端 JSON 图结构，校验 `Code` 唯一性，初始化 `Version=1`，存入 PG `jsonb`。
2. **流程列表 & 详情 (List & Get)**
	- **后端**：实现 `ListProcessDefinitions` 和 `GetProcessDefinition`。
	- **逻辑**：列表页不返回 `structure` 大字段；详情页返回完整 JSON 供前端 ReactFlow 回显。
3. **发布/停用 (Toggle)**
	- **后端**：实现 `UpdateProcessStatus`。
	- **逻辑**：简单的状态开关，控制该流程是否允许发起新工单。

> **里程碑**：前端可以在设计器里画图并保存，管理后台能看到流程列表。

#### 第二阶段：引擎启动 (Instantiation) —— "把车造出来"

**目标**：实现核心调度器 `Scheduler` 的启动逻辑，生成第一个 Token。

1. **发起工单 (Start Instance)**
	- **后端**：实现 `CreateProcessInstance` 接口。
	- **逻辑**：
		1. 锁定流程定义的快照 ID。
		2. 保存表单数据到 `Variables`。
		3. 初始化 `ActiveTokens` 指向 `StartEvent` 节点。
		4. **关键**：调用引擎 `Scheduler.MoveToken`，将 Token 从 Start 节点自动推送到第一个 UserTask 节点。
2. **任务生成 (Task Production)**
	- **后端**：实现 `Scheduler` 内部逻辑。
	- **逻辑**：当 Token 到达 `UserTask` 节点时，解析 `assignee` 规则（MVP 先做最简单的：指定固定 `user_id` 或从表单变量取），在 `process_tasks` 表插入一条 `PENDING` 记录。

> **里程碑**：调用“发起接口”后，数据库里多了一条 Instance 记录，且 Task 表里多了一条待办任务。

#### 第三阶段：流转与决策 (Execution) —— "让车跑起来"

**目标**：实现任务的完成和网关的判断。

1. **查询待办 (My Todo)**
	- **后端**：实现 `ListUserTasks` 接口。
	- **逻辑**：查询 `process_tasks` 表，筛选 `assignee = 当前用户 AND status = 'PENDING'`。
2. **审批任务 (Approve/Reject)**
	- **后端**：实现 `CompleteTask` 接口。
	- **逻辑**：
		1. 更新 Task 状态为 `APPROVED` 或 `REJECTED`。
		2. 写入审批意见 `Comment`。
		3. **关键**：触发 `Scheduler.MoveToken` 继续往下走。
3. **网关解析 (Gateway Routing)**
	- **后端**：实现 `GatewayParser`。
	- **逻辑**：当 Token 到达网关时，利用 `expr` 库计算连线上的条件（如 `amount > 500`），决定 Token 走向哪个分支。

> **里程碑**：用户可以在“待办”里看到任务，点击同意后，任务消失，流程流转到下一个人；如果遇到网关，能根据金额自动分流。

#### 第四阶段：感知与反馈 (Observation) —— "仪表盘"

**目标**：完善用户侧的查询和通知。

1. **我申请的 / 我已处理 (Applied/Handled)**
	- **后端**：实现 `ListMyInstances` 和 `ListMyHandledTasks`。
	- **逻辑**：简单的 SQL 查询（利用 `submitter_id` 和历史 Task 记录）。
2. **工单详情与日志 (Detail & Log)**
	- **后端**：实现 `GetInstanceDetail`。
	- **逻辑**：返回实例基本信息 + 审批轨迹（将 `process_tasks` 按时间轴排列返回，展示谁在什么时候批了什么）。
3. **基础通知 (Notification)**
	- **后端**：实现 `Notifier` 接口的 Log 实现（MVP）。
	- **逻辑**：在任务生成时，打印日志 "发送飞书通知给用户 X"。后续接入飞书 SDK 时替换此实现即可。