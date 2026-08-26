import { useEffect, useState } from "react";
import { createOpenClient, OPEN_SERVER_BASE } from "@minimall-open/sdk-js";

export default function App() {
  const [health, setHealth] = useState("loading…");
  useEffect(() => {
    createOpenClient()
      .health()
      .then((d) => setHealth(JSON.stringify(d, null, 2)))
      .catch((e: Error) => setHealth(e.message));
  }, []);
  return (
    <main style={{ fontFamily: "system-ui", padding: "2rem", maxWidth: 800, margin: "0 auto" }}>
      <h1>MiniMall Open Platform</h1>
      <p style={{ color: "#64748b" }}>开发者门户壳 · 应用管理 / 密钥 / 文档（待建设）</p>
      <p><code>open-server</code>: {OPEN_SERVER_BASE}</p>
      <pre style={{ background: "#f8fafc", padding: "1rem", borderRadius: 8 }}>{health}</pre>
    </main>
  );
}
