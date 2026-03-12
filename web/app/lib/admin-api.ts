export type AuditEvent = {
  timestamp: string;
  kind: string;
  request: {
    id?: string;
    server: string;
    tool: string;
    method: string;
    arguments?: Record<string, unknown>;
  };
  decision?: {
    action?: string;
    reason?: string;
    rule_name?: string;
    allowed?: boolean;
  };
  upstream?: {
    target?: string;
    outcome?: string;
    http_status_code?: number;
    latency?: number;
    error?: string;
    forwarded?: boolean;
  };
};

export type ApprovalRequest = {
  id: string;
  status: string;
  server: string;
  tool: string;
  method: string;
  reason: string;
  rule_name?: string;
  request_id?: string;
  created_at: string;
  created_by?: string;
  resolved_at?: string;
  resolved_by?: string;
  resolution?: string;
};

export type PolicyStatus = {
  configured: boolean;
  ruleCount: number;
  source: string;
  summary: string;
};

export type ApiState<T> = {
  data: T | null;
  error: string | null;
};

const apiBase = process.env.AGENTFENCE_API_BASE ?? "http://127.0.0.1:8080";

async function getJSON<T>(path: string): Promise<ApiState<T>> {
  try {
    const response = await fetch(`${apiBase}${path}`, {
      cache: "no-store",
      headers: {
        Accept: "application/json",
      },
    });

    if (!response.ok) {
      const body = await response.text();
      return {
        data: null,
        error: body || `Request failed with status ${response.status}`,
      };
    }

    const data = (await response.json()) as T;
    return { data, error: null };
  } catch (error) {
    return {
      data: null,
      error: error instanceof Error ? error.message : "Unknown request failure",
    };
  }
}

export async function getRecentAuditEvents(limit = 20) {
  return getJSON<{ items: AuditEvent[]; available: boolean }>(
    `/api/admin/audit?limit=${limit}`,
  );
}

export async function getPendingApprovals() {
  return getJSON<{ items: ApprovalRequest[]; available: boolean }>(
    "/api/admin/approvals",
  );
}

export async function getPolicyStatus() {
  return getJSON<PolicyStatus>("/api/admin/policy");
}
