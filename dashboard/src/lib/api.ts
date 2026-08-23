/**
 * Tarak REST API Client with automatic Bearer token injection
 */

export const getAuthToken = (): string | null => {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("tarak_token");
};

export const setAuthToken = (token: string): void => {
  if (typeof window === "undefined") return;
  localStorage.setItem("tarak_token", token);
};

export const removeAuthToken = (): void => {
  if (typeof window === "undefined") return;
  localStorage.removeItem("tarak_token");
};

export async function tarakFetch<T = any>(
  path: string,
  options: RequestInit = {}
): Promise<{ data: T | null; error: string | null; status: number }> {
  const token = getAuthToken();
  const headers = new Headers(options.headers || {});

  if (!headers.has("Content-Type") && options.body && typeof options.body === "string") {
    headers.set("Content-Type", "application/json");
  }

  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  try {
    const res = await fetch(path, {
      ...options,
      headers,
    });

    const status = res.status;

    if (!res.ok) {
      const text = await res.text();
      let errorMsg = `HTTP ${status}: ${res.statusText}`;
      try {
        const json = JSON.parse(text);
        if (json.message) errorMsg = json.message;
      } catch {}
      return { data: null, error: errorMsg, status };
    }

    const contentType = res.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      const data = await res.json();
      return { data, error: null, status };
    }

    const text = (await res.text()) as unknown as T;
    return { data: text, error: null, status };
  } catch (err: any) {
    return { data: null, error: err?.message || "Network request failed", status: 0 };
  }
}
