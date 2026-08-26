export const OPEN_SERVER_BASE =
  import.meta.env?.VITE_OPEN_SERVER_BASE?.replace(/\/$/, "") ??
  import.meta.env?.VITE_OPEN_BFF_BASE?.replace(/\/$/, "") ??
  "http://127.0.0.1:8092";

/** @deprecated use OPEN_SERVER_BASE */
export const OPEN_BFF_BASE = OPEN_SERVER_BASE;

export function createOpenClient(baseUrl = OPEN_SERVER_BASE) {
  return {
    async health(): Promise<Record<string, unknown>> {
      const res = await fetch(`${baseUrl}/open/v1/health`);
      if (!res.ok) throw new Error(`open-server health failed: ${res.status}`);
      return res.json() as Promise<Record<string, unknown>>;
    },
    async ping(): Promise<Record<string, unknown>> {
      const res = await fetch(`${baseUrl}/open/v1/ping`);
      if (!res.ok) throw new Error(`open-server ping failed: ${res.status}`);
      return res.json() as Promise<Record<string, unknown>>;
    },
  };
}
