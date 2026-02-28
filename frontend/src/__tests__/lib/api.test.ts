import {
  fetchHealth,
  fetchWorkflows,
  fetchWorkflowRuns,
  fetchTaskRuns,
  fetchWorkers,
  createWorkflow,
  triggerWorkflow,
  getWsUrl,
} from "@/lib/api";

const mockFetch = jest.fn();
global.fetch = mockFetch;

function mockResponse<T>(body: T, ok = true, status = 200) {
  mockFetch.mockResolvedValueOnce({
    ok,
    status,
    json: () => Promise.resolve(body),
  });
}

beforeEach(() => {
  mockFetch.mockReset();
});

describe("fetchHealth", () => {
  it("returns health data on success", async () => {
    mockResponse({ status: "ok", service: "scheduler" });
    const result = await fetchHealth();
    expect(result).toEqual({ status: "ok", service: "scheduler" });
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/healthz"),
      undefined
    );
  });

  it("throws on non-ok response", async () => {
    mockResponse(null, false, 503);
    await expect(fetchHealth()).rejects.toThrow("API error 503");
  });
});

describe("fetchWorkflows", () => {
  it("fetches workflows with default offset and limit", async () => {
    mockResponse([]);
    await fetchWorkflows();
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/workflows?offset=0&limit=20"),
      undefined
    );
  });

  it("passes custom offset and limit", async () => {
    mockResponse([]);
    await fetchWorkflows(10, 50);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/workflows?offset=10&limit=50"),
      undefined
    );
  });

  it("returns workflow list", async () => {
    const workflows = [
      {
        id: "w1",
        name: "daily-etl",
        description: "ETL job",
        schedule_cron: "0 0 * * *",
        is_active: true,
        created_at: "2024-01-01T00:00:00Z",
      },
    ];
    mockResponse(workflows);
    const result = await fetchWorkflows();
    expect(result).toEqual(workflows);
  });
});

describe("fetchWorkflowRuns", () => {
  it("fetches without status filter by default", async () => {
    mockResponse([]);
    await fetchWorkflowRuns();
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/workflow-runs"),
      undefined
    );
    expect(mockFetch.mock.calls[0][0]).not.toContain("?status=");
  });

  it("appends status query param when provided", async () => {
    mockResponse([]);
    await fetchWorkflowRuns("running");
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/workflow-runs?status=running"),
      undefined
    );
  });
});

describe("fetchTaskRuns", () => {
  it("fetches without status filter by default", async () => {
    mockResponse([]);
    await fetchTaskRuns();
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/task-runs"),
      undefined
    );
    expect(mockFetch.mock.calls[0][0]).not.toContain("?status=");
  });

  it("appends status query param when provided", async () => {
    mockResponse([]);
    await fetchTaskRuns("failed");
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/task-runs?status=failed"),
      undefined
    );
  });
});

describe("fetchWorkers", () => {
  it("fetches worker list", async () => {
    const workers = [
      {
        id: "worker-1",
        hostname: "host-a",
        last_heartbeat: "2024-01-01T00:00:00Z",
        status: "active" as const,
      },
    ];
    mockResponse(workers);
    const result = await fetchWorkers();
    expect(result).toEqual(workers);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/workers"),
      undefined
    );
  });
});

describe("createWorkflow", () => {
  it("sends POST with JSON body", async () => {
    const payload = {
      name: "test",
      description: "desc",
      schedule_cron: "* * * * *",
      is_active: true,
    };
    const created = { ...payload, id: "new-id", created_at: "2024-01-01T00:00:00Z" };
    mockResponse(created);
    const result = await createWorkflow(payload);
    expect(result).toEqual(created);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/workflows"),
      expect.objectContaining({
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
    );
  });
});

describe("triggerWorkflow", () => {
  it("sends POST to trigger endpoint", async () => {
    const run = {
      id: "run-1",
      workflow_id: "w1",
      status: "pending" as const,
      started_at: "2024-01-01T00:00:00Z",
    };
    mockResponse(run);
    const result = await triggerWorkflow("w1");
    expect(result).toEqual(run);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/workflows/w1/trigger"),
      expect.objectContaining({ method: "POST" })
    );
  });
});

describe("getWsUrl", () => {
  const originalEnv = process.env;

  afterEach(() => {
    process.env = originalEnv;
  });

  it("converts http to ws", () => {
    process.env.NEXT_PUBLIC_API_URL = "http://localhost:8080";
    const url = getWsUrl();
    expect(url).toBe("ws://localhost:8080/ws/updates");
  });

  it("converts https to wss", () => {
    process.env.NEXT_PUBLIC_API_URL = "https://api.example.com";
    const url = getWsUrl();
    expect(url).toBe("wss://api.example.com/ws/updates");
  });
});
