/**
 * Tests for /api/toml route handler.
 *
 * Coverage:
 *  - Valid requests (asset found, asset not in TOML, upstream parse error)
 *  - SSRF protection  (non-https protocol, blocked hosts, private IP ranges, path traversal attempts)
 *  - Rate limiting    (30 req/min per IP; resets after window)
 *  - Upstream errors  (non-2xx response, network failure / fetch throw)
 *  - Missing / malformed query parameters
 *
 * Implementation note:
 *  The route module keeps a module-level `rateLimitMap`.  To guarantee test
 *  isolation the rate-limiting suite re-imports the module via
 *  `vi.resetModules()` + dynamic `import()` before each test so every test
 *  starts with an empty map.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { NextRequest } from "next/server";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Build a NextRequest for the /api/toml endpoint.
 *
 * @param params - Query-string parameters to include.
 * @param ip     - Value for the x-forwarded-for header (defaults to "1.2.3.4").
 */
function makeTomlRequest(
  params: Record<string, string>,
  ip = "1.2.3.4"
): NextRequest {
  const url = new URL("http://localhost/api/toml");
  for (const [key, value] of Object.entries(params)) {
    url.searchParams.set(key, value);
  }
  return new NextRequest(url.toString(), {
    headers: { "x-forwarded-for": ip },
  });
}

/** Minimal valid TOML with one CURRENCIES entry. */
const VALID_TOML_TEXT = `
[DOCUMENTATION]
ORG_NAME = "Acme Corp"
ORG_LOGO = "https://acme.example.com/logo.png"

[[CURRENCIES]]
code = "USDC"
issuer = "GISSUER1234567890ABCDEF"
name = "USD Coin"
desc = "Dollar-backed stablecoin"
image = "https://acme.example.com/usdc.png"
`;

/** TOML that has no CURRENCIES array. */
const EMPTY_TOML_TEXT = `
[DOCUMENTATION]
ORG_NAME = "Bare Corp"
`;

/** TOML text that is syntactically invalid. */
const INVALID_TOML_TEXT = `[broken toml = {{{`;

/** A valid public URL that should pass SSRF checks. */
const VALID_URL = "https://acme.example.com/.well-known/stellar.toml";

/** Shared valid query params for a happy-path request. */
const VALID_PARAMS = {
  url: VALID_URL,
  code: "USDC",
  issuer: "GISSUER1234567890ABCDEF",
};

// ---------------------------------------------------------------------------
// Shared fetch-mock helpers
// ---------------------------------------------------------------------------

/**
 * Stub `globalThis.fetch` to return a successful TOML response.
 *
 * Uses `mockImplementation` (not `mockResolvedValue`) so that each call
 * receives a **new** `Response` instance — `Response` bodies can only be
 * consumed once, so reusing the same object across calls causes subsequent
 * reads to throw.
 */
function stubFetchSuccess(body: string): void {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(body, {
          status: 200,
          headers: { "Content-Type": "text/plain" },
        })
      )
    )
  );
}

/**
 * Stub `globalThis.fetch` to return an HTTP error response.
 */
function stubFetchError(status: number): void {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(new Response("", { status }))
  );
}

/**
 * Stub `globalThis.fetch` to throw a network-level error.
 */
function stubFetchThrows(error: Error): void {
  vi.stubGlobal("fetch", vi.fn().mockRejectedValue(error));
}

// ---------------------------------------------------------------------------
// Parameter validation
// ---------------------------------------------------------------------------

describe("GET /api/toml – parameter validation", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns 400 when the 'url' parameter is missing", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({ code: "USDC", issuer: "GISSUER" });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/url/i);
  });

  it("returns 400 when the 'code' parameter is missing", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({ url: VALID_URL, issuer: "GISSUER" });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/code|issuer/i);
  });

  it("returns 400 when the 'issuer' parameter is missing", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({ url: VALID_URL, code: "USDC" });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/code|issuer/i);
  });

  it("returns 400 when the URL is not parseable", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "not-a-url",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/invalid url/i);
  });
});

// ---------------------------------------------------------------------------
// SSRF protection
// ---------------------------------------------------------------------------

describe("GET /api/toml – SSRF protection", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("rejects HTTP (non-HTTPS) URLs", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "http://example.com/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/https/i);
  });

  it("rejects localhost by name", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "https://localhost/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/internal/i);
  });

  it("rejects 127.0.0.1", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "https://127.0.0.1/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/internal/i);
  });

  it("rejects 0.0.0.0", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "https://0.0.0.0/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/internal/i);
  });

  it("rejects IPv6 loopback [::1]", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "https://[::1]/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/internal/i);
  });

  it("rejects 10.x.x.x (RFC-1918 private range)", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "https://10.0.0.1/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/internal/i);
  });

  it("rejects 192.168.x.x (RFC-1918 private range)", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "https://192.168.1.100/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/internal/i);
  });

  it("rejects 172.x.x.x (RFC-1918 private range)", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "https://172.16.0.1/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/internal/i);
  });

  it("rejects 169.254.x.x (link-local / APIPA range)", async () => {
    const { GET } = await import("./route");
    const req = makeTomlRequest({
      url: "https://169.254.169.254/.well-known/stellar.toml",
      code: "USDC",
      issuer: "GISSUER",
    });
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/internal/i);
  });
});

// ---------------------------------------------------------------------------
// Rate limiting
// ---------------------------------------------------------------------------

describe("GET /api/toml – rate limiting", () => {
  /**
   * Each test re-imports the module so that the module-level rateLimitMap
   * starts empty and fake-timer control is precise.
   */

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it("allows requests under the 30-per-minute limit", async () => {
    // Fake only Date so AbortSignal.timeout() (which uses real setTimeout) still works.
    vi.useFakeTimers({ toFake: ["Date"] });
    stubFetchSuccess(VALID_TOML_TEXT);
    const { GET } = await import("./route");

    for (let i = 0; i < 30; i++) {
      const req = makeTomlRequest(VALID_PARAMS, "5.5.5.5");
      const res = await GET(req);
      // All 30 should succeed (200) — not be rate-limited
      expect(res.status, `Request ${i + 1} should not be rate limited`).toBe(200);
    }
  });

  it("blocks the 31st request from the same IP within the same window", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    stubFetchSuccess(VALID_TOML_TEXT);
    const { GET } = await import("./route");
    const IP = "9.9.9.9";

    // Exhaust the 30-request allowance
    for (let i = 0; i < 30; i++) {
      await GET(makeTomlRequest(VALID_PARAMS, IP));
    }

    const res = await GET(makeTomlRequest(VALID_PARAMS, IP));
    expect(res.status).toBe(429);
    const body = await res.json();
    expect(body.error).toMatch(/too many requests/i);
  });

  it("does not rate-limit a different IP", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    stubFetchSuccess(VALID_TOML_TEXT);
    const { GET } = await import("./route");

    // Exhaust allowance for IP_A
    for (let i = 0; i < 30; i++) {
      await GET(makeTomlRequest(VALID_PARAMS, "11.11.11.11"));
    }

    // IP_B should still be allowed
    const res = await GET(makeTomlRequest(VALID_PARAMS, "22.22.22.22"));
    expect(res.status).toBe(200);
  });

  it("resets the counter after the 1-minute window expires", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    stubFetchSuccess(VALID_TOML_TEXT);
    const { GET } = await import("./route");
    const IP = "33.33.33.33";

    // Exhaust the allowance
    for (let i = 0; i < 30; i++) {
      await GET(makeTomlRequest(VALID_PARAMS, IP));
    }

    // Advance time past the 60-second window
    vi.advanceTimersByTime(61_000);

    // The counter should have reset; the request should succeed
    const res = await GET(makeTomlRequest(VALID_PARAMS, IP));
    expect(res.status).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// Happy-path – valid request
// ---------------------------------------------------------------------------

describe("GET /api/toml – valid request", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns 200 with asset metadata when the currency is found in the TOML", async () => {
    stubFetchSuccess(VALID_TOML_TEXT);
    const { GET } = await import("./route");

    const req = makeTomlRequest(VALID_PARAMS);
    const res = await GET(req);

    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.code).toBe("USDC");
    expect(body.issuer).toBe("GISSUER1234567890ABCDEF");
    expect(body.name).toBe("USD Coin");
    expect(body.description).toBe("Dollar-backed stablecoin");
    expect(body.imageUrl).toBe("https://acme.example.com/usdc.png");
    expect(body.orgName).toBe("Acme Corp");
    expect(body.orgLogo).toBe("https://acme.example.com/logo.png");
  });

  it("sets Cache-Control to 24 hours on success", async () => {
    stubFetchSuccess(VALID_TOML_TEXT);
    const { GET } = await import("./route");

    const res = await GET(makeTomlRequest(VALID_PARAMS));
    expect(res.headers.get("Cache-Control")).toContain("max-age=86400");
  });

  it("returns code and issuer when the asset is not found in CURRENCIES", async () => {
    stubFetchSuccess(VALID_TOML_TEXT);
    const { GET } = await import("./route");

    const req = makeTomlRequest({
      url: VALID_URL,
      code: "NOTFOUND",
      issuer: "GISSUER1234567890ABCDEF",
    });
    const res = await GET(req);

    expect(res.status).toBe(200);
    const body = await res.json();
    // Metadata should contain the requested code/issuer with no enrichment
    expect(body.code).toBe("NOTFOUND");
    expect(body.issuer).toBe("GISSUER1234567890ABCDEF");
    expect(body.name).toBeUndefined();
  });

  it("matches currency code case-insensitively", async () => {
    stubFetchSuccess(VALID_TOML_TEXT);
    const { GET } = await import("./route");

    const req = makeTomlRequest({
      url: VALID_URL,
      code: "usdc", // lowercase
      issuer: "GISSUER1234567890ABCDEF",
    });
    const res = await GET(req);

    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.name).toBe("USD Coin");
  });

  it("returns metadata with only code and issuer when TOML has no CURRENCIES", async () => {
    stubFetchSuccess(EMPTY_TOML_TEXT);
    const { GET } = await import("./route");

    const res = await GET(makeTomlRequest(VALID_PARAMS));
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.code).toBe("USDC");
    expect(body.issuer).toBe("GISSUER1234567890ABCDEF");
    expect(body.name).toBeUndefined();
  });

  it("returns code/issuer and 1-hour Cache-Control on invalid TOML (graceful fallback)", async () => {
    stubFetchSuccess(INVALID_TOML_TEXT);
    const { GET } = await import("./route");

    const res = await GET(makeTomlRequest(VALID_PARAMS));
    // The route catches parse errors and returns a graceful empty metadata
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.code).toBe("USDC");
    expect(res.headers.get("Cache-Control")).toContain("max-age=3600");
  });
});

// ---------------------------------------------------------------------------
// Upstream errors
// ---------------------------------------------------------------------------

describe("GET /api/toml – upstream errors", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns the upstream status code when fetch responds with a non-2xx status", async () => {
    stubFetchError(503);
    const { GET } = await import("./route");

    const res = await GET(makeTomlRequest(VALID_PARAMS));
    expect(res.status).toBe(503);
    const body = await res.json();
    expect(body.error).toMatch(/503/);
  });

  it("returns 404 propagated from upstream", async () => {
    stubFetchError(404);
    const { GET } = await import("./route");

    const res = await GET(makeTomlRequest(VALID_PARAMS));
    expect(res.status).toBe(404);
  });

  it("returns graceful fallback metadata when fetch throws a network error", async () => {
    stubFetchThrows(new TypeError("network failure"));
    const { GET } = await import("./route");

    const res = await GET(makeTomlRequest(VALID_PARAMS));
    // Network errors are caught; the handler returns metadata with only code/issuer
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.code).toBe("USDC");
    expect(body.issuer).toBe("GISSUER1234567890ABCDEF");
    expect(res.headers.get("Cache-Control")).toContain("max-age=3600");
  });

  it("returns graceful fallback metadata when the request times out", async () => {
    stubFetchThrows(new DOMException("The operation was aborted.", "AbortError"));
    const { GET } = await import("./route");

    const res = await GET(makeTomlRequest(VALID_PARAMS));
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.code).toBe("USDC");
    expect(res.headers.get("Cache-Control")).toContain("max-age=3600");
  });
});
