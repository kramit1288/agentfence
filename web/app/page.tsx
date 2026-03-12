const sections = [
  {
    title: "Policy",
    body: "Allow, deny, and approval-required decisions will be managed here.",
  },
  {
    title: "Audit",
    body: "Durable request and response records will surface here for operators.",
  },
  {
    title: "Approvals",
    body: "Queued human approvals will live in this administrative surface.",
  },
];

export default function Home() {
  return (
    <main className="shell">
      <section className="hero">
        <p className="eyebrow">AgentFence</p>
        <h1>Secure MCP tool access before it reaches production systems.</h1>
        <p className="lede">
          This initial UI scaffold establishes the operator dashboard surface
          without embedding business logic yet.
        </p>
      </section>

      <section className="grid">
        {sections.map((section) => (
          <article className="card" key={section.title}>
            <h2>{section.title}</h2>
            <p>{section.body}</p>
          </article>
        ))}
      </section>
    </main>
  );
}
