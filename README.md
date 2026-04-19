## Dawnix 流程引擎

一个基于 Go 的流程引擎后端，支持流程定义、实例流转、任务审批和邮件服务节点。

### 快速启动

1. 准备依赖：Postgres、Redis。
	 - 可使用 Docker 启动 Postgres（用户名 `root`，密码 `123`）：

```bash
docker run -d \
	--name dawnix-pg \
	-e POSTGRES_USER=root \
	-e POSTGRES_PASSWORD=123 \
	-e POSTGRES_DB=root \
	-p 5432:5432 \
	postgres:16
```

2. 默认情况下邮件节点已关闭，无需配置 `SMTP_TOKEN`。如果手动启用邮件特性，再补充 `SMTP_TOKEN`。
3. 启动服务：

```bash
go run ./cmd/server
```

默认端口：`8080`

配置说明：默认读取 `configs/dev.yaml`，`local.env` 与系统环境变量可覆盖 `EMAIL_SERVICE_ENABLED`、`SMTP_TOKEN`、`SMTP_EMAIL`、`REDIS_ADDR`、`DB_DSN`。
数据库连接串建议使用 URI 形式，例如：`postgres://root:123@localhost:5432/root?sslmode=disable&TimeZone=Asia/Shanghai`

### 基础检查

```bash
go build ./...
go test ./...
```

说明：`./client` 的邮件测试依赖 `SMTP_TOKEN`，未配置时会失败。

### 已实现功能

- 流程定义管理：创建、列表、详情、删除
- 流程实例管理：按流程编码启动实例、列表、详情、删除
- 任务管理：任务列表、任务详情、审批完成（agree/reject）
- 调度引擎：支持 start/end、user_task、fork/join、xor、inclusive；email_service 作为可选插件默认关闭
- 网关路由：
	- XOR：命中唯一条件分支或默认分支
	- Inclusive：支持多分支命中，未命中走默认分支
- 异步邮件节点：已插件化，默认关闭，启用后才会投递 Redis 队列并启动 worker
- 分层架构：api / service / biz / data，仓储与事务边界已落地

### 将要实现功能

- 接入真实鉴权与用户上下文（替换任务列表中的硬编码用户）
- 完善任务列表 scope（如 all_pending、all_completed 等）
- 增加更完整的端到端测试与网关场景测试
- 提供更完整的接口文档和流程建模示例
- 补充分环境配置（dev/test/prod）与密钥管理规范