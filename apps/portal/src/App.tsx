import { FormEvent, useEffect, useMemo, useState } from "react";
import { createOpenClient, OPEN_SERVER_BASE } from "@minimall-open/sdk-js";

type ApiItem = {
  name: string;
  method: string;
  path: string;
  summary: string;
  params: Array<{ name: string; required?: boolean; description: string; example: string }>;
};

const API_CATALOG: ApiItem[] = [
  {
    name: "health",
    method: "GET",
    path: "/open/v1/health",
    summary: "健康检查",
    params: [],
  },
  {
    name: "apis",
    method: "GET",
    path: "/open/v1/apis",
    summary: "API 目录",
    params: [],
  },
  {
    name: "sign",
    method: "POST",
    path: "/open/v1/debug/sign",
    summary: "签名预览",
    params: [
      { name: "appKey", required: true, description: "应用 Key", example: "demo_app_key" },
      { name: "timestamp", required: true, description: "时间戳", example: "1720000000" },
      { name: "path", required: true, description: "请求路径", example: "/open/v1/products/get" },
    ],
  },
  {
    name: "execute",
    method: "POST",
    path: "/open/v1/debug/execute",
    summary: "执行调试请求",
    params: [
      { name: "apiName", required: true, description: "接口名", example: "taobao.product.get" },
      { name: "method", required: true, description: "HTTP 方法", example: "GET" },
      { name: "path", required: true, description: "请求路径", example: "/open/v1/products/get" },
    ],
  },
];

function pretty(value: unknown) {
  return JSON.stringify(value, null, 2);
}

export default function App() {
  const [selected, setSelected] = useState(API_CATALOG[0]);
  const [appKey, setAppKey] = useState("demo_app_key");
  const [appSecret, setAppSecret] = useState("demo_app_secret");
  const [params, setParams] = useState<Record<string, string>>({
    apiName: "taobao.product.get",
    method: "GET",
    path: "/open/v1/products/get",
    appKey: "demo_app_key",
    timestamp: String(Math.floor(Date.now() / 1000)),
  });
  const [response, setResponse] = useState<string>('{"status":"idle"}');
  const [loading, setLoading] = useState(false);
  const [health, setHealth] = useState("loading...");

  useEffect(() => {
    createOpenClient()
      .health()
      .then((d) => setHealth(pretty(d)))
      .catch((e: Error) => setHealth(e.message));
  }, []);

  const activeParams = useMemo(() => selected.params, [selected]);

  function updateParam(name: string, value: string) {
    setParams((prev) => ({ ...prev, [name]: value }));
  }

  async function runRequest(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    try {
      const payload = Object.fromEntries(new FormData(e.currentTarget).entries());
      const res = await fetch(`${OPEN_SERVER_BASE}${selected.path}`, {
        method: selected.method,
        headers: { "Content-Type": "application/json" },
        body: selected.method === "GET" ? undefined : JSON.stringify(payload),
      });
      const data = await res.json();
      setResponse(pretty({ ok: res.ok, status: res.status, data }));
    } catch (err) {
      setResponse(pretty({ ok: false, error: err instanceof Error ? err.message : String(err) }));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="portal">
      <section className="hero">
        <div>
          <div className="eyebrow">MiniMall Open Platform</div>
          <h1>API 调试平台</h1>
          <p className="lede">
            对照淘宝开放平台的调试工具，我们先把接口目录、参数编辑、签名预览、请求执行和响应查看这条闭环跑通。
          </p>
        </div>
        <div className="hero-card">
          <div className="card-title">Open Server</div>
          <code>{OPEN_SERVER_BASE}</code>
          <pre>{health}</pre>
        </div>
      </section>

      <section className="grid">
        <aside className="panel">
          <div className="panel-title">API 目录</div>
          <div className="api-list">
            {API_CATALOG.map((api) => (
              <button
                key={api.name}
                className={api.name === selected.name ? "api-item active" : "api-item"}
                onClick={() => setSelected(api)}
                type="button"
              >
                <strong>{api.name}</strong>
                <span>{api.method} {api.path}</span>
                <small>{api.summary}</small>
              </button>
            ))}
          </div>
        </aside>

        <section className="panel">
          <div className="panel-title">请求编辑器</div>
          <form className="form" onSubmit={runRequest}>
            <label>
              <span>应用 Key</span>
              <input value={appKey} onChange={(e) => setAppKey(e.target.value)} name="appKey" />
            </label>
            <label>
              <span>应用 Secret</span>
              <input value={appSecret} onChange={(e) => setAppSecret(e.target.value)} name="appSecret" />
            </label>
            <label>
              <span>接口名</span>
              <input value={params.apiName ?? ""} onChange={(e) => updateParam("apiName", e.target.value)} name="apiName" />
            </label>
            <label>
              <span>方法</span>
              <input value={selected.method} readOnly name="method" />
            </label>
            <label>
              <span>路径</span>
              <input value={params.path ?? selected.path} onChange={(e) => updateParam("path", e.target.value)} name="path" />
            </label>
            <label>
              <span>时间戳</span>
              <input value={params.timestamp ?? ""} onChange={(e) => updateParam("timestamp", e.target.value)} name="timestamp" />
            </label>

            <div className="param-table">
              <div className="param-head">参数模板</div>
              {activeParams.length === 0 ? (
                <div className="param-empty">当前接口没有额外参数</div>
              ) : activeParams.map((item) => (
                <label key={item.name} className="param-row">
                  <span>
                    {item.name}
                    {item.required ? " *" : ""}
                    <small>{item.description}</small>
                  </span>
                  <input
                    name={item.name}
                    placeholder={item.example}
                    value={params[item.name] ?? ""}
                    onChange={(e) => updateParam(item.name, e.target.value)}
                  />
                </label>
              ))}
            </div>

            <div className="actions">
              <button type="submit" disabled={loading}>
                {loading ? "执行中..." : "发送请求"}
              </button>
              <button
                type="button"
                className="ghost"
                onClick={() => setParams({ ...params, timestamp: String(Math.floor(Date.now() / 1000)) })}
              >
                刷新时间戳
              </button>
            </div>
          </form>
        </section>

        <section className="panel">
          <div className="panel-title">响应结果</div>
          <pre className="response">{response}</pre>
          <div className="note">
            下一步可以把这里替换成更接近淘宝工具的参数分组、签名串展示和 curl 复制。
          </div>
        </section>
      </section>
    </main>
  );
}
