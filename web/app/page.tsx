import Link from "next/link";

import { EmptyState } from "./components/empty-state";
import { ErrorState } from "./components/error-state";
import {
  getPendingApprovals,
  getPolicyStatus,
  getRecentAuditEvents,
} from "./lib/admin-api";

export default async function Home() {
  const [auditState, approvalsState, policyState] = await Promise.all([
    getRecentAuditEvents(8),
    getPendingApprovals(),
    getPolicyStatus(),
  ]);

  const auditItems = auditState.data?.items ?? [];
  const approvalItems = approvalsState.data?.items ?? [];
  const policy = policyState.data;

  return (
    <div className="page-stack">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Overview</p>
          <h2>Operate the MCP gateway without leaving unsafe gaps.</h2>
          <p className="lede">
            Review recent request decisions, resolve pending approvals, and
            inspect whether policy enforcement is active in the current runtime.
          </p>
        </div>
        <div className="hero-meta">
          <span>{auditItems.length} recent audit events</span>
          <span>{approvalItems.length} pending approvals</span>
          <span>{policy?.configured ? "policy loaded" : "policy missing"}</span>
        </div>
      </section>

      <section className="stats-grid">
        <article className="metric-card">
          <span className="metric-label">Policy</span>
          <strong>{policy?.configured ? "Configured" : "Not configured"}</strong>
          <p>{policy?.summary ?? "Policy status is unavailable."}</p>
        </article>
        <article className="metric-card">
          <span className="metric-label">Recent audit</span>
          <strong>{auditItems.length}</strong>
          <p>Most recent audit records returned by the backend.</p>
        </article>
        <article className="metric-card">
          <span className="metric-label">Pending approvals</span>
          <strong>{approvalItems.length}</strong>
          <p>Requests still waiting on a human decision.</p>
        </article>
      </section>

      <section className="panel-grid">
        <article className="panel">
          <div className="panel-head">
            <h3>Recent Audit Events</h3>
            <Link href="/audit">View all</Link>
          </div>
          {auditState.error ? (
            <ErrorState title="Audit unavailable" detail={auditState.error} />
          ) : auditItems.length === 0 ? (
            <EmptyState
              title="No audit events yet"
              detail="Recent gateway activity will appear here once requests are recorded."
            />
          ) : (
            <ul className="event-list">
              {auditItems.map((event) => (
                <li key={`${event.kind}-${event.timestamp}-${event.request.id ?? "none"}`}>
                  <div>
                    <strong>{event.request.tool}</strong>
                    <p>
                      {event.kind} on {event.request.server}
                    </p>
                  </div>
                  <span>{new Date(event.timestamp).toLocaleString()}</span>
                </li>
              ))}
            </ul>
          )}
        </article>

        <article className="panel">
          <div className="panel-head">
            <h3>Pending Approvals</h3>
            <Link href="/approvals">Open queue</Link>
          </div>
          {approvalsState.error ? (
            <ErrorState
              title="Approvals unavailable"
              detail={approvalsState.error}
            />
          ) : approvalItems.length === 0 ? (
            <EmptyState
              title="No pending approvals"
              detail="Approval-required tool calls will appear here when the queue is active."
            />
          ) : (
            <ul className="approval-list">
              {approvalItems.slice(0, 6).map((item) => (
                <li key={item.id}>
                  <div>
                    <strong>{item.tool}</strong>
                    <p>
                      {item.server} · {item.reason}
                    </p>
                  </div>
                  <code>{item.id}</code>
                </li>
              ))}
            </ul>
          )}
        </article>
      </section>
    </div>
  );
}
