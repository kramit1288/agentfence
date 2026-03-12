type ErrorStateProps = {
  title: string;
  detail: string;
};

export function ErrorState({ title, detail }: ErrorStateProps) {
  return (
    <div className="error-state" role="alert">
      <strong>{title}</strong>
      <p>{detail}</p>
    </div>
  );
}
