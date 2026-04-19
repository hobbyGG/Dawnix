# Dawnix Docker 初始化指南

## 环境要求

- Docker 已安装
- PostgreSQL 容器 `pg_dawnix` 运行在 localhost:5432
- Redis 容器 `redis_dawnix` 运行在 localhost:6379

## 当前状态

已初始化的 PostgreSQL 数据库包含以下表：

```
✓ dawnix_process_definition    - 流程定义
✓ dawnix_process_instance      - 流程实例
✓ dawnix_execution             - 执行令牌
✓ dawnix_process_task          - 用户任务
```

所有表都已创建索引并配置了外键关系。

## 快速开始

### 1. 启动后端服务

```bash
cd /Users/umep/prj/Dawnix
go run cmd/server/main.go
```

服务将在 `http://localhost:8080/api/v1` 上运行。

### 2. 启动前端服务

```bash
cd /Users/umep/prj/dawnix_fe
npm run dev
```

访问 `http://localhost:5173`

### 3. 验证连接

```bash
# 检查后端 API
curl http://localhost:8080/api/v1/process/definitions

# 检查 PostgreSQL 连接
docker exec -i pg_dawnix psql -U root -d root -c "SELECT COUNT(*) FROM dawnix_process_definition;"
```

## 数据库连接信息

- **主机**: localhost
- **端口**: 5432
- **数据库**: root
- **用户**: root
- **密码**: 123
- **时区**: Asia/Shanghai

## 常用 Docker 命令

```bash
# 查看容器状态
docker ps

# 连接到 PostgreSQL
docker exec -it pg_dawnix psql -U root -d root

# 查看表
docker exec -it pg_dawnix psql -U root -d root -c "\dt"

# 查看表结构
docker exec -it pg_dawnix psql -U root -d root -c "\d dawnix_process_definition"

# 清空所有数据（谨慎使用）
docker exec -i pg_dawnix psql -U root -d root << 'EOF'
DROP TABLE IF EXISTS dawnix_process_task CASCADE;
DROP TABLE IF EXISTS dawnix_execution CASCADE;
DROP TABLE IF EXISTS dawnix_process_instance CASCADE;
DROP TABLE IF EXISTS dawnix_process_definition CASCADE;
EOF
```

## 故障排除

### 连接被拒绝 (Connection refused)

确保 PostgreSQL 容器正在运行：
```bash
docker ps | grep pg_dawnix
```

如果未运行，启动它：
```bash
docker start pg_dawnix
```

### "database root does not exist"

连接到 postgres 数据库并创建 root 数据库：
```bash
docker exec -i pg_dawnix psql -U root -d postgres -c "CREATE DATABASE root;"
```

### 表不存在

重新运行本文档中的 SQL 初始化脚本。

## 环境变量配置

在 `local.env` 中配置（或使用默认值）：

```env
# Database (default: postgres://root:123@localhost:5432/root?sslmode=disable&TimeZone=Asia/Shanghai)
DATABASE_DSN=postgres://root:123@localhost:5432/root?sslmode=disable&TimeZone=Asia/Shanghai

# Redis (default: 127.0.0.1:16379)
REDIS_ADDR=127.0.0.1:16379

# Email service plugin is disabled by default.
# Enable only when you need the email node and worker.
EMAIL_SERVICE_ENABLED=false

# SMTP Token (required only when email service is enabled)
SMTP_TOKEN=your_smtp_token_here
```

## 生产建议

- 修改默认密码
- 启用 SSL 连接 (sslmode=require)
- 配置备份策略
- 设置适当的连接池参数
- 启用容器资源限制
