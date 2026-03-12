import { EmptyState } from "../components/empty-state";
import { ErrorState } from "../components/error-state";
import { getRecentAuditEvents } from "../lib/admin-api";

export default async function AuditPage() {
  const auditState = await getRecentAuditEvents(50);
  const items = auditState.data?.items ?? [];

  return (
    <section className="page-stack">
      <div className="page-header">
        <div>
          <p className="eyebrow">Audit</p>
          <h2>Recent audit events</h2>
        </div>
        <p className="section-copy">
          This feed shows the latest redacted decision and upstream activity.
        </p>
      </div>

      {auditState.error ? (
        <ErrorState title="Audit API error" detail={auditState.error} />
      ) : items.length === 0 ? (
        <EmptyState
          title="No recent events"
          detail="Gateway activity will appear here after requests are handled."
        />
      ) : (
        <div className="table-wrap">
          <table className="data-table">
            <thead>
              <tr>
                <th>Time</th>
                <th>Kind</th>
                <th>Server</th>
                <th>Tool</th>
                <th>Outcome</th>
              </tr>
            </thead>
            <tbody>
              {items.map((event) => {
                const outcome =
                  event.kind === "upstream.call"
                    ? event.upstream?.outcome ?? "unknown"
                    : event.decision?.action ?? "unknown";

                return (
                  <tr
                    key={`${event.kind}-${event.timestamp}-${event.request.id ?? "none"}`}
                  >
                    <td>{new Date(event.timestamp).toLocaleString()}</td>
                    <td>{event.kind}</td>
                    <td>{event.request.server}</td>
                    <td>{event.request.tool}</td>
                    <td>{outcome}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
