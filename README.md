# mini-mall-open

开放平台 monorepo：**Go `server/`** + 开发者门户 + OpenAPI 规范 + SDK 示例。

架构定稿见 [mini-mall-services/docs/architecture-plan-a.md](https://github.com/minimall-labs/mini-mall-services/blob/main/docs/architecture-plan-a.md)。

## 结构

```
mini-mall-open/
├── server/              # open server（Go → mini-mall-gateway）
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

## 北向流量

```
ISV / 第三方  ──►  open server (:8092)  ──►  mini-mall-gateway (:8080)  ──►  *-service
开发者门户    ──►  portal (:5300)     ──►  open server
```

## 开发

```bash
pnpm install
pnpm dev              # 开发者门户
pnpm dev:server       # open server（Go）

cd server && go run .
```

| 服务 | 地址 |
|------|------|
| open server | http://127.0.0.1:8092 |
| portal | http://127.0.0.1:5300 |

健康检查：`GET http://127.0.0.1:8092/open/v1/health`

调试接口：

- `GET /open/v1/apis`：API 目录
- `POST /open/v1/debug/sign`：签名预览
- `POST /open/v1/debug/execute`：调试执行（目前先回显请求体）

门户首页现在就是一个轻量 API 调试台，可以选择接口、编辑参数并查看响应。

环境变量：

| 变量 | 默认 |
|------|------|
| `PORT` | `8092` |
| `MINIMALL_GATEWAY_BASE_URL` | `http://127.0.0.1:8080` |

## 相关

- 领域服务：[minimall-labs/mini-mall-services](https://github.com/minimall-labs/mini-mall-services)
- C 端：[minimall-labs/mini-mall-consumer](https://github.com/minimall-labs/mini-mall-consumer)
- B 端：[minimall-labs/mini-mall-workbench](https://github.com/minimall-labs/mini-mall-workbench)
