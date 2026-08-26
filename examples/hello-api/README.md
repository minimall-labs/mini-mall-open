# hello-api

最小 ISV 调用示例（shell）：

```bash
curl -s http://127.0.0.1:8092/bff/open/health | jq .
curl -s http://127.0.0.1:8092/bff/open/ping | jq .
```

或使用 `@minimall-open/sdk-js` 的 `createOpenClient().health()`。
