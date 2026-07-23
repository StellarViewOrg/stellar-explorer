/**
 * Tests for /api/stellar-expert route handler.
 *
 * Coverage:
 *  - Valid requests (happy path, response forwarding, Cache-Control header)
 *  - Path allow-list validation (valid paths, invalid paths, path traversal)
 *  - Rate limiting  (60 req/min per IP; resets after window; per-IP isolation)
 *  - Upstream errors (non-2xx, network failure / fetch throw)
 *  - Missing query parameter
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
 * Build a NextRequest for the /api/stellar-expert endpoint.
 *
 * @param path - The `path` query-string value (or omitted to test missing).
 * @param ip   - Value for the x-forwarded-for header (defaults to "1.2.3.4").
 */
function makeStellarExpertRequest(
  path: string | null,
  ip = "1.2.3.4"
): NextRequest {
  const url = new URL("http://localhost/api/stellar-expert");
  if (path !== null) {
    url.searchParams.set("path", path);
  }
  return new NextRequest(url.toString(), {
    headers: { "x-forwarded-for": ip },
  });
}

/** A sample valid path that matches the allow-list regex. */
const VALID_PATH = "explorer/public/asset/USDC-GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN";

/** A valid upstream JSON response body. */
const UPSTREAM_PAYLOAD = { _links: {}, records: [{ id: 1 }] };

// ---------------------------------------------------------------------------
// Shared fetch-mock helpers
// ---------------------------------------------------------------------------

/**
 * Stub `globalThis.fetch` to return a successful JSON response.
 *
 * Uses `mockImplementation` (not `mockResolvedValue`) so that each call
 * receives a **new** `Response` instance — `Response` bodies can only be
 * consumed once, so reusing the same object across calls causes subsequent
 * reads to throw.
 */
function stubFetchSuccess(body: unknown = UPSTREAM_PAYLOAD): void {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(body), {
          status: 200,
          headers: { "Content-Type": "application/json" },
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
// Missing parameter
// ---------------------------------------------------------------------------

describe("GET /api/stellar-expert – missing parameter", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns 400 when the 'path' parameter is absent", async () => {
    const { GET } = await import("./route");
    const req = makeStellarExpertRequest(null);
    const res = await GET(req);
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/path/i);
  });
});

// ---------------------------------------------------------------------------
// Path allow-list (acts as the SSRF guard for this proxy route)
// ---------------------------------------------------------------------------

describe("GET /api/stellar-expert – path allow-list / SSRF guard", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  const VALID_PATHS = [
    "explorer/public/asset/USDC-GABCDE",
    "explorer/testnet/account/GABCDE",
    "explorer/futurenet/tx/abc123",
    "explorer/public/asset/XLM",
    // path with dots and colons (e.g., asset code + issuer separated by colon)
    "explorer/public/asset/USDC:GABCDE",
  ];

  const INVALID_PATHS = [
    // Wrong network segment
    "explorer/mainnet/asset/USDC",
    // Missing network segment
    "explorer/asset/USDC",
    // Does not start with "explorer/"
    "api/public/asset/USDC",
    "public/asset/USDC",
    // Empty remaining path after network
    "explorer/public/",
    // Absolute path
    "/explorer/public/asset/USDC",
  ];

  for (const validPath of VALID_PATHS) {
    it(`allows valid path: "${validPath}"`, async () => {
      stubFetchSuccess();
      const { GET } = await import("./route");
      vi.resetModules(); // re-reset so next iteration gets fresh map
      const res = await GET(makeStellarExpertRequest(validPath));
      expect(res.status).toBe(200);
    });
  }

  for (const invalidPath of INVALID_PATHS) {
    it(`rejects invalid path: "${invalidPath}"`, async () => {
      const { GET } = await import("./route");
      vi.resetModules();
      const res = await GET(makeStellarExpertRequest(invalidPath));
      expect(res.status).toBe(400);
      const body = await res.json();
      expect(body.error).toMatch(/invalid path/i);
    });
  }

  it("rejects path traversal via '..'", async () => {
    const { GET } = await import("./route");
    const res = await GET(
      makeStellarExpertRequest("explorer/public/../../../etc/passwd")
    );
    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/invalid path/i);
  });

  it("rejects path traversal encoded as URL segments", async () => {
    const { GET } = await import("./route");
    // Even without percent-encoding, a literal ".." embedded in the path is blocked
    const res = await GET(
      makeStellarExpertRequest("explorer/public/asset/../../../secret")
    );
    expect(res.status).toBe(400);
  });
});

// ---------------------------------------------------------------------------
// Rate limiting
// ---------------------------------------------------------------------------

describe("GET /api/stellar-expert – rate limiting", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it("allows up to 60 requests per minute from the same IP", async () => {
    // Fake only Date so AbortSignal.timeout() (which uses real setTimeout) still works.
    vi.useFakeTimers({ toFake: ["Date"] });
    stubFetchSuccess();
    const { GET } = await import("./route");

    for (let i = 0; i < 60; i++) {
      const res = await GET(makeStellarExpertRequest(VALID_PATH, "5.5.5.5"));
      expect(res.status, `Request ${i + 1} should not be rate limited`).toBe(200);
    }
  });

  it("blocks the 61st request within the same window", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    stubFetchSuccess();
    const { GET } = await import("./route");
    const IP = "9.9.9.9";

    for (let i = 0; i < 60; i++) {
      await GET(makeStellarExpertRequest(VALID_PATH, IP));
    }

    const res = await GET(makeStellarExpertRequest(VALID_PATH, IP));
    expect(res.status).toBe(429);
    const body = await res.json();
    expect(body.error).toMatch(/too many requests/i);
  });

  it("does not rate-limit a different IP", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    stubFetchSuccess();
    const { GET } = await import("./route");

    for (let i = 0; i < 60; i++) {
      await GET(makeStellarExpertRequest(VALID_PATH, "11.11.11.11"));
    }

    const res = await GET(makeStellarExpertRequest(VALID_PATH, "22.22.22.22"));
    expect(res.status).toBe(200);
  });

  it("resets the counter after the 1-minute window expires", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    stubFetchSuccess();
    const { GET } = await import("./route");
    const IP = "33.33.33.33";

    for (let i = 0; i < 60; i++) {
      await GET(makeStellarExpertRequest(VALID_PATH, IP));
    }

    // Advance past the 60-second window
    vi.advanceTimersByTime(61_000);

    const res = await GET(makeStellarExpertRequest(VALID_PATH, IP));
    expect(res.status).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// Happy-path – valid request
// ---------------------------------------------------------------------------

describe("GET /api/stellar-expert – valid request", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns 200 with the upstream JSON body", async () => {
    stubFetchSuccess(UPSTREAM_PAYLOAD);
    const { GET } = await import("./route");

    const res = await GET(makeStellarExpertRequest(VALID_PATH));
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body).toEqual(UPSTREAM_PAYLOAD);
  });

  it("forwards the request to the correct stellar.expert URL", async () => {
    const fetchSpy = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(UPSTREAM_PAYLOAD), { status: 200 })
    );
    vi.stubGlobal("fetch", fetchSpy);
    const { GET } = await import("./route");

    await GET(makeStellarExpertRequest(VALID_PATH));

    expect(fetchSpy).toHaveBeenCalledOnce();
    const calledUrl: string = fetchSpy.mock.calls[0][0] as string;
    expect(calledUrl).toBe(`https://stellar.expert/api/${VALID_PATH}`);
  });

  it("sets Cache-Control to 5 minutes on success", async () => {
    stubFetchSuccess();
    const { GET } = await import("./route");

    const res = await GET(makeStellarExpertRequest(VALID_PATH));
    expect(res.headers.get("Cache-Control")).toContain("max-age=300");
  });

  it("uses 'unknown' IP when x-forwarded-for header is absent", async () => {
    stubFetchSuccess();
    const { GET } = await import("./route");

    // Build request without x-forwarded-for header
    const url = new URL("http://localhost/api/stellar-expert");
    url.searchParams.set("path", VALID_PATH);
    const req = new NextRequest(url.toString());
    const res = await GET(req);
    expect(res.status).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// Upstream errors
// ---------------------------------------------------------------------------

describe("GET /api/stellar-expert – upstream errors", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("propagates the upstream HTTP status on non-2xx responses", async () => {
    stubFetchError(503);
    const { GET } = await import("./route");

    const res = await GET(makeStellarExpertRequest(VALID_PATH));
    expect(res.status).toBe(503);
    const body = await res.json();
    expect(body.error).toMatch(/503/);
    expect(body.fallback).toBe(true);
  });

  it("propagates 404 from upstream", async () => {
    stubFetchError(404);
    const { GET } = await import("./route");

    const res = await GET(makeStellarExpertRequest(VALID_PATH));
    expect(res.status).toBe(404);
    const body = await res.json();
    expect(body.fallback).toBe(true);
  });

  it("returns 502 with fallback flag when fetch throws a network error", async () => {
    stubFetchThrows(new TypeError("network failure"));
    const { GET } = await import("./route");

    const res = await GET(makeStellarExpertRequest(VALID_PATH));
    expect(res.status).toBe(502);
    const body = await res.json();
    expect(body.error).toMatch(/failed to fetch/i);
    expect(body.fallback).toBe(true);
  });

  it("returns 502 with fallback flag when the request times out", async () => {
    stubFetchThrows(new DOMException("The operation was aborted.", "AbortError"));
    const { GET } = await import("./route");

    const res = await GET(makeStellarExpertRequest(VALID_PATH));
    expect(res.status).toBe(502);
    const body = await res.json();
    expect(body.fallback).toBe(true);
  });
});
