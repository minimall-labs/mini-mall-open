# mini-mall-open

开放平台 monorepo：**open-bff** + 开发者门户 + OpenAPI 规范 + SDK 示例。

## 结构

```
mini-mall-open/
├── bff/                 # open-bff（Spring Boot → mini-mall Gateway）
├── apps/
│   └── portal/          # 开发者控制台壳（:5300）
├── packages/
│   ├── openapi/         # OpenAPI 规范源文件
│   └── sdk-js/          # JavaScript SDK 壳
└── examples/
    └── hello-api/       # 最小调用示例
```

## 北向流量

```
ISV / 第三方  ──►  open-bff (:8092)  ──►  mini-mall Gateway (:8080)  ──►  *-service
开发者门户    ──►  portal (:5300)   ──►  open-bff
```

## 开发

```bash
pnpm install
pnpm dev              # 开发者门户
pnpm dev:bff          # open-bff

cd bff && mvn spring-boot:run
```

| 服务 | 地址 |
|------|------|
| open-bff | http://127.0.0.1:8092 |
| portal | http://127.0.0.1:5300 |
| OpenAPI 文档 | http://127.0.0.1:8092/swagger-ui.html |

## 相关

- 后端：[minimall-labs/mini-mall](https://github.com/minimall-labs/mini-mall)
- C 端：[minimall-labs/mini-mall-consumer](https://github.com/minimall-labs/mini-mall-consumer)
- B 端：[minimall-labs/mini-mall-workbench](https://github.com/minimall-labs/mini-mall-workbench)
