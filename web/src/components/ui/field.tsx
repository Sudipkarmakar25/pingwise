import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes, TextareaHTMLAttributes } from "react";

const controlClass = (error?: string) =>
  `mt-1.5 block w-full rounded-md border bg-surface px-3 text-sm text-ink placeholder:text-muted/70 focus:outline-2 focus:outline-offset-0 focus:outline-accent disabled:bg-canvas disabled:text-muted ${
    error ? "border-down" : "border-line"
  }`;

type Common = { label: string; name: string; hint?: ReactNode; error?: string; className?: string };

// FieldShell draws the label, hint and error around a control and links them for screen readers
function FieldShell({
  label,
  id,
  hint,
  error,
  className = "",
  children,
}: Omit<Common, "name"> & { id: string; children: (describedBy?: string) => ReactNode }) {
  const describedBy = error ? `${id}-error` : hint ? `${id}-hint` : undefined;
  return (
    <div className={className}>
      <label htmlFor={id} className="block text-sm font-medium text-ink">
        {label}
      </label>
      {children(describedBy)}
      {error ? (
        <p id={`${id}-error`} className="mt-1.5 text-sm text-down-ink">
          {error}
        </p>
      ) : hint ? (
        <p id={`${id}-hint`} className="mt-1.5 text-sm text-muted">
          {hint}
        </p>
      ) : null}
    </div>
  );
}

export function Field({ label, name, hint, error, id, className, ...props }: Common & InputHTMLAttributes<HTMLInputElement>) {
  const inputId = id ?? name;
  return (
    <FieldShell label={label} id={inputId} hint={hint} error={error} className={className}>
      {(describedBy) => (
        <input
          id={inputId}
          name={name}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className={`h-10 ${controlClass(error)}`}
          {...props}
        />
      )}
    </FieldShell>
  );
}

export function SelectField({
  label,
  name,
  hint,
  error,
  id,
  className,
  children,
  ...props
}: Common & SelectHTMLAttributes<HTMLSelectElement>) {
  const selectId = id ?? name;
  return (
    <FieldShell label={label} id={selectId} hint={hint} error={error} className={className}>
      {(describedBy) => (
        <select
          id={selectId}
          name={name}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className={`h-10 ${controlClass(error)}`}
          {...props}
        >
          {children}
        </select>
      )}
    </FieldShell>
  );
}

export function TextAreaField({
  label,
  name,
  hint,
  error,
  id,
  className,
  ...props
}: Common & TextareaHTMLAttributes<HTMLTextAreaElement>) {
  const areaId = id ?? name;
  return (
    <FieldShell label={label} id={areaId} hint={hint} error={error} className={className}>
      {(describedBy) => (
        <textarea
          id={areaId}
          name={name}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className={`py-2 font-mono ${controlClass(error)}`}
          {...props}
        />
      )}
    </FieldShell>
  );
}
