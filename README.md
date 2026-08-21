# 数据源配置包校验与发布平台

这是一个完全离线可运行的配置发布系统。管理员上传含多个数据源定义的 JSON 配置包后，系统先解析字段位置，再执行类型、字段、名称、环境、引用和循环依赖校验；只有经过不同人员批准的无错误配置包才能在一个事务中整体或选择性发布。

## 模块职责

- `cmd/server`：依赖装配、HTTP 服务和优雅停机。
- `internal/domain`：配置类型、配置、版本、状态、影响和审计模型。
- `internal/application`：上传、校验、预览、批准、原子发布、回滚、删除、敏感导出和影响分析用例。
- `internal/service`：JSON 行号解析、字段规范校验、依赖图和差异计算。
- `internal/repository`：事务与存储接口；`memory` 提供离线实现，`postgres` 提供 pgx 连接和迁移入口。
- `internal/transport/http`：Gin `/api/v1` 路由、稳定错误结构和本地身份头。
- `internal/platform`：AES-GCM、本地附件、内存通知、本地健康检查和标识生成器。
- `migrations`：可重复执行的 PostgreSQL 迁移和不覆盖已有值的演示类型。
- `api/openapi`：OpenAPI 3.0 接口契约。
- `web`：Vue 3、TypeScript、Vite 和 Pinia 中文管理端。

## 本地启动

需要 Go 1.24+、Node.js 20+ 和可选的 PostgreSQL 16。默认服务使用离线内存仓储，因此不启动数据库也能运行和测试。

```bash
cp .env.example .env
go run ./cmd/server
cd web
npm install
npm run dev
```

后端监听 `:8080`，前端监听 `:5173`。前端开发代理会把 `/api` 转到本地后端。健康检查为 `GET /healthz`，就绪检查为 `GET /readyz`。

## 配置

| 变量 | 默认值 | 用途 |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | HTTP 监听地址 |
| `DATABASE_URL` | 本地 PostgreSQL URL | pgx 持久化连接 |
| `MASTER_KEY` | 本地开发固定值 | 32 字节 AES-GCM 主密钥，生产必须替换 |
| `ATTACHMENT_DIR` | `./data/attachments` | 受控本地附件目录 |
| `MAX_UPLOAD_BYTES` | `4194304` | JSON 上传上限 |
| `REQUEST_TIMEOUT` | `5s` | 请求及本地适配器超时 |
| `SHUTDOWN_TIMEOUT` | `10s` | 优雅停机窗口 |

平台使用 `X-Actor-ID` 和 `X-Actor-Roles` 作为离线可验证身份适配器。角色包括 `platform_admin`、`config_editor`、`approver`、`publisher` 和 `secret_exporter`。上传和发布同时要求 `Idempotency-Key`。

## 迁移与演示数据

迁移位于 `migrations/`，可按文件名顺序通过 `psql` 重复执行：

```bash
psql "$DATABASE_URL" -f migrations/001_initial.sql
psql "$DATABASE_URL" -f migrations/002_seed_types.sql
```

种子脚本使用 `ON CONFLICT DO NOTHING`，不会覆盖已有配置类型。服务的内存模式会登记 `postgres`、`http_api` 和 `derived` 三种相同语义的本地演示类型。

## 接口示例

上传配置包：

```bash
curl -X POST http://localhost:8080/api/v1/bundles \
  -H 'X-Actor-ID: local-admin' \
  -H 'X-Actor-Roles: platform_admin,config_editor' \
  -H 'Idempotency-Key: upload-001' \
  -F 'file=@examples/bundle.json;type=application/json'
```

校验、批准和发布分别调用：

```text
POST /api/v1/bundles/{id}/validate
POST /api/v1/bundles/{id}/approve
POST /api/v1/bundles/{id}/publish
```

所有错误响应包含稳定 `code`、中文 `message`、可选 `fields` 和 `request_id`。列表请求采用 `page`、`size`、`sort` 和白名单筛选字段，非法筛选不会传入仓储层。

## 状态和事务规则

配置状态按 `draft -> pending -> published -> withdrawn -> archived` 的受控边转换，撤回配置可再次发布，归档配置不可恢复。所有更新带 `revision` 做乐观并发控制。

配置包依次经过 `uploaded`、`validated`、`approved` 和 `published`。校验有错误时进入 `rejected`；上传者不能批准自己的配置包。发布在单个事务快照内创建或更新配置、写版本、写审计并保存幂等结果，任一步失败都保持零写入。被其他配置引用的配置不能删除。

敏感字段在持久化前使用 AES-GCM 加密。普通导出始终遮罩；只有具备 `secret_exporter` 角色、声明用途且授权未过期时才能解密下载，并写入审计事件。

## 测试与验证

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
cd web && npm test
cd web && npm run build
```

当前基线已覆盖 JSON 行号、字段与环境规范、依赖环、上传幂等、职责分离、并发发布重放、乐观锁与引用删除、版本回滚、敏感导出和并发影响分析。运行期附件、依赖缓存、前端产物和任何密钥均由 `.gitignore` 排除。

