/** 北向 API 基址：统一经 Gateway 进入 Open BFF。 */
export const GATEWAY_BASE =
  import.meta.env?.VITE_GATEWAY_BASE?.replace(/\/$/, "") ?? "http://127.0.0.1:8080";

export const OPEN_SERVER_BASE =
  import.meta.env?.VITE_OPEN_SERVER_BASE?.replace(/\/$/, "") ??
  import.meta.env?.VITE_OPEN_BFF_BASE?.replace(/\/$/, "") ??
  GATEWAY_BASE;

/** @deprecated use OPEN_SERVER_BASE */
export const OPEN_BFF_BASE = OPEN_SERVER_BASE;

export function createOpenClient(baseUrl = OPEN_SERVER_BASE) {
  return {
    async health(): Promise<Record<string, unknown>> {
      const res = await fetch(`${baseUrl}/open/v1/health`);
      if (!res.ok) throw new Error(`open-server health failed: ${res.status}`);
      return res.json() as Promise<Record<string, unknown>>;
    },
    async apis(): Promise<Record<string, unknown>> {
      const res = await fetch(`${baseUrl}/open/v1/apis`);
      if (!res.ok) throw new Error(`open-server apis failed: ${res.status}`);
      return res.json() as Promise<Record<string, unknown>>;
    },
    async ping(): Promise<Record<string, unknown>> {
      const res = await fetch(`${baseUrl}/open/v1/ping`);
      if (!res.ok) throw new Error(`open-server ping failed: ${res.status}`);
      return res.json() as Promise<Record<string, unknown>>;
    },
  };
}
