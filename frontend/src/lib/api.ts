const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type Status = "pending" | "running" | "success" | "failed";
export type WorkerStatus = "active" | "inactive";

export interface Workflow {
  id: string;
  name: string;
  description: string;
  schedule_cron: string;
  is_active: boolean;
  created_at: string;
}

export interface WorkflowRun {
  id: string;
  workflow_id: string;
  status: Status;
  started_at: string;
  finished_at?: string;
}

export interface TaskRun {
  id: string;
  workflow_run_id: string;
  task_id: string;
  status: Status;
  attempt: number;
  started_at: string;
  finished_at?: string;
  logs: string;
}

export interface Worker {
  id: string;
  hostname: string;
  last_heartbeat: string;
  status: WorkerStatus;
}

export interface HealthResponse {
  status: string;
  service: string;
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, options);
  if (!res.ok) throw new Error(`API error ${res.status}: ${path}`);
  return res.json() as Promise<T>;
}

export async function fetchHealth(): Promise<HealthResponse> {
  return apiFetch<HealthResponse>("/healthz");
}

export async function fetchWorkflows(offset = 0, limit = 20): Promise<Workflow[]> {
  return apiFetch<Workflow[]>(`/workflows?offset=${offset}&limit=${limit}`);
}

export async function createWorkflow(body: Omit<Workflow, "id" | "created_at">): Promise<Workflow> {
  return apiFetch<Workflow>("/workflows", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

export async function triggerWorkflow(id: string): Promise<WorkflowRun> {
  return apiFetch<WorkflowRun>(`/workflows/${id}/trigger`, { method: "POST" });
}

export async function fetchWorkflowRuns(status?: Status): Promise<WorkflowRun[]> {
  const qs = status ? `?status=${status}` : "";
  return apiFetch<WorkflowRun[]>(`/workflow-runs${qs}`);
}

export async function fetchTaskRuns(status?: Status): Promise<TaskRun[]> {
  const qs = status ? `?status=${status}` : "";
  return apiFetch<TaskRun[]>(`/task-runs${qs}`);
}

export async function fetchWorkers(): Promise<Worker[]> {
  return apiFetch<Worker[]>("/workers");
}

export function getWsUrl(): string {
  return (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080")
    .replace(/^http/, "ws") + "/ws/updates";
}
