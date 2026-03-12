import { EmptyState } from "../components/empty-state";
import { ErrorState } from "../components/error-state";
import { getPendingApprovals } from "../lib/admin-api";

export default async function ApprovalsPage() {
  const approvalsState = await getPendingApprovals();
  const items = approvalsState.data?.items ?? [];

  return (
    <section className="page-stack">
      <div className="page-header">
        <div>
          <p className="eyebrow">Approvals</p>
          <h2>Pending approval queue</h2>
        </div>
        <p className="section-copy">
          Review requests blocked by policy until a human decision is recorded.
        </p>
      </div>

      {approvalsState.error ? (
        <ErrorState title="Approval API error" detail={approvalsState.error} />
      ) : items.length === 0 ? (
        <EmptyState
          title="No pending approvals"
          detail="The queue is currently empty."
        />
      ) : (
        <div className="table-wrap">
          <table className="data-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Server</th>
                <th>Tool</th>
                <th>Created</th>
                <th>Reason</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id}>
                  <td>
                    <code>{item.id}</code>
                  </td>
                  <td>{item.server}</td>
                  <td>{item.tool}</td>
                  <td>{new Date(item.created_at).toLocaleString()}</td>
                  <td>{item.reason}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
