# mini-mall-open

开放平台 monorepo：**Go `server/`** + 开发者门户 + OpenAPI 规范 + SDK 示例。

平台架构见 [mini-mall-services/docs/architecture.md](https://github.com/minimall-labs/mini-mall-services/blob/main/docs/architecture.md) · 方案 A：[architecture-plan-a.md](https://github.com/minimall-labs/mini-mall-services/blob/main/docs/architecture-plan-a.md)。

## 结构

```
mini-mall-open/
├── server/              # open server（Go BFF）
├── apps/
│   └── portal/          # 开发者控制台壳（:5300）
├── packages/
│   ├── openapi/         # OpenAPI 规范源文件
│   └── sdk-js/          # JavaScript SDK 壳
└── examples/
    └── hello-api/       # 最小调用示例
```

## 边界

| 属于本仓库 | 不属于（另立内部系统） |
|------------|------------------------|
| 开发者门户、AppKey 管理壳 | 平台运营 Admin |
| 对外 Open API（签名、频控） | 商家工作台 |
| OpenAPI + SDK + 示例 | |

## 北向 / 南向

```
北向：ISV / SDK / 开发者门户  →  Gateway (:8080)  →  Open BFF (:8092)

南向：Open BFF  →  Gateway  →  *-service（领域读接口规划中）
```

门户默认经 Gateway 访问 Open API（`OPEN_SERVER_BASE` / SDK 默认 `http://127.0.0.1:8080/open/v1`）。

## 开发

```bash
pnpm install
pnpm dev              # 开发者门户
pnpm dev:server       # open server（Go）

cd server && go run .
```

| 服务 | 地址 |
|------|------|
| Gateway（北向入口） | http://127.0.0.1:8080 |
| open server（BFF） | http://127.0.0.1:8092 |
| portal | http://127.0.0.1:5300 |

健康检查（经 Gateway）：`GET http://127.0.0.1:8080/open/v1/health`

调试接口：

- `GET /open/v1/apis`：API 目录
- `POST /open/v1/debug/sign`：签名预览
- `POST /open/v1/debug/execute`：调试执行（目前先回显请求体）
- `GET /metrics`：Prometheus 风格指标

观测能力：

- 每个请求会自动生成或透传 `X-Request-Id` 和 `X-Trace-Id`
- 日志输出为结构化 JSON，包含路由、状态码、耗时和 trace 信息
- BFF 转发到 Gateway 时会透传 trace 头，方便串起链路

门户首页现在就是一个轻量 API 调试台，可以选择接口、编辑参数并查看响应。

环境变量：

| 变量 | 默认 |
|------|------|
| `PORT` | `8092` |
| `MINIMALL_GATEWAY_BASE_URL` | `http://127.0.0.1:8080` |

## 相关

- 领域服务：[minimall-labs/mini-mall-services](https://github.com/minimall-labs/mini-mall-services)
- 网关：[minimall-labs/mini-mall-gateway](https://github.com/minimall-labs/mini-mall-gateway)
- C 端：[minimall-labs/mini-mall-consumer](https://github.com/minimall-labs/mini-mall-consumer)
- B 端：[minimall-labs/mini-mall-workbench](https://github.com/minimall-labs/mini-mall-workbench)
