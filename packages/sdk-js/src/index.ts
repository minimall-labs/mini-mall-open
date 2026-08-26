export const OPEN_BFF_BASE =
  import.meta.env?.VITE_OPEN_BFF_BASE?.replace(/\/$/, "") ?? "http://127.0.0.1:8092";

export function createOpenClient(baseUrl = OPEN_BFF_BASE) {
  return {
    async health(): Promise<Record<string, unknown>> {
      const res = await fetch(`${baseUrl}/bff/open/health`);
      if (!res.ok) throw new Error(`open-bff health failed: ${res.status}`);
      return res.json() as Promise<Record<string, unknown>>;
    },
    async ping(): Promise<Record<string, unknown>> {
      const res = await fetch(`${baseUrl}/bff/open/ping`);
      if (!res.ok) throw new Error(`open-bff ping failed: ${res.status}`);
      return res.json() as Promise<Record<string, unknown>>;
    },
  };
}
