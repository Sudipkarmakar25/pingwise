"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import type { ReactNode } from "react";
import { Alert } from "@/components/ui/alert";
import { Button, ButtonLink } from "@/components/ui/button";
import { Field, SelectField, TextAreaField } from "@/components/ui/field";
import { api } from "@/lib/api";
import { intervalLabel } from "@/lib/format";
import type { Service } from "@/lib/types";
import { useFormSubmit } from "@/lib/use-form-submit";

const INTERVALS = [1, 2, 3, 5, 10, 15, 30, 60];
const MAX_HEADERS = 20;
const HEADER_NAME = /^[A-Za-z0-9-]+$/;

// none: no auth. set: entering a header and token. keep: an existing token stays. remove: drop it.
type AuthMode = "none" | "set" | "keep" | "remove";
type HeaderRow = { id: number; name: string; value: string };

type FormState = {
  name: string;
  url: string;
  method: Service["method"];
  expectedStatus: string;
  timeoutSeconds: string;
  intervalMinutes: string;
  failureThreshold: string;
  authMode: AuthMode;
  authHeaderName: string;
  authToken: string;
  headers: HeaderRow[];
  requestBody: string;
};

let nextRowId = 1;

function initialState(service?: Service): FormState {
  return {
    name: service?.name ?? "",
    url: service?.url ?? "",
    method: service?.method ?? "GET",
    expectedStatus: String(service?.expected_status ?? 200),
    timeoutSeconds: String(service?.timeout_seconds ?? 10),
    intervalMinutes: String(service?.interval_minutes ?? 1),
    failureThreshold: String(service?.failure_threshold ?? 2),
    authMode: service?.auth_token_masked ? "keep" : "none",
    authHeaderName: service?.auth_header_name ?? "Authorization",
    authToken: "",
    headers: Object.entries(service?.custom_headers ?? {}).map(([name, value]) => ({ id: nextRowId++, name, value })),
    requestBody: service?.request_body ?? "",
  };
}

// validate catches mistakes before the request, with messages that name the problem
function validate(form: FormState): Record<string, string> {
  const errors: Record<string, string> = {};
  if (!form.name.trim()) errors.name = "Enter a name for this service.";
  if (!/^https?:\/\/\S+$/i.test(form.url.trim())) errors.url = "Enter a full URL starting with http:// or https://.";
  if (form.authMode === "set" || form.authMode === "keep") {
    if (!HEADER_NAME.test(form.authHeaderName)) {
      errors.auth_header_name = "Use letters, numbers and hyphens only, for example Authorization or X-Api-Key.";
    }
  }
  if (form.authMode === "set" && !form.authToken) errors.auth_token = "Enter the token, or turn off authentication.";
  const seen = new Set<string>();
  for (const row of form.headers) {
    if (!row.name && !row.value) continue;
    const key = row.name.toLowerCase();
    if (!HEADER_NAME.test(row.name)) {
      errors.custom_headers = `"${row.name || "(empty)"}" isn't a valid header name. Use letters, numbers and hyphens.`;
    } else if (seen.has(key)) {
      errors.custom_headers = `The header ${row.name} is listed twice.`;
    } else if ((form.authMode === "set" || form.authMode === "keep") && key === form.authHeaderName.toLowerCase()) {
      errors.custom_headers = `${row.name} is already used for authentication. Remove it from the custom headers.`;
    }
    seen.add(key);
  }
  return errors;
}

function payload(form: FormState, service?: Service) {
  const body: Record<string, unknown> = {
    name: form.name.trim(),
    url: form.url.trim(),
    method: form.method,
    expected_status: Number(form.expectedStatus),
    timeout_seconds: Number(form.timeoutSeconds),
    interval_minutes: Number(form.intervalMinutes),
    failure_threshold: Number(form.failureThreshold),
    custom_headers: Object.fromEntries(form.headers.filter((h) => h.name || h.value).map((h) => [h.name, h.value])),
  };
  // only POST sends a body; on edit an empty value clears a body saved earlier
  const requestBody = form.method === "POST" ? form.requestBody : "";
  if (service || requestBody) body.request_body = requestBody;

  if (form.authMode === "set") {
    body.auth_header_name = form.authHeaderName;
    body.auth_token = form.authToken;
  } else if (form.authMode === "keep" && form.authHeaderName !== service?.auth_header_name) {
    body.auth_header_name = form.authHeaderName;
  } else if (form.authMode === "remove") {
    body.auth_token = "";
  }
  return body;
}

export function ServiceForm({ service }: { service?: Service }) {
  const router = useRouter();
  const [form, setForm] = useState(() => initialState(service));
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const { pending, formError, fieldErrors, submit, setError } = useFormSubmit({
    name: "Enter a name, up to 100 characters.",
    url: "Enter a full URL starting with http:// or https://, up to 2,000 characters.",
    method: "Choose GET, POST or HEAD.",
    expected_status: "Enter a status code between 100 and 599.",
    timeout_seconds: "Enter a timeout between 1 and 30 seconds.",
    interval_minutes: "Choose one of the listed intervals.",
    failure_threshold: "Choose between 1 and 5 failures.",
    auth_header_name: "Use up to 100 characters.",
    auth_token: "The token can be up to 4,000 characters.",
    custom_headers: "Use up to 20 headers, with names up to 100 characters and values up to 2,000.",
    request_body: "The request body can be up to 10,000 characters.",
  });
  const errors = { ...fieldErrors, ...clientErrors };
  const set = <K extends keyof FormState>(key: K, value: FormState[K]) => setForm((f) => ({ ...f, [key]: value }));
  const setHeader = (id: number, change: Partial<HeaderRow>) =>
    set("headers", form.headers.map((h) => (h.id === id ? { ...h, ...change } : h)));

  return (
    <form
      noValidate
      onSubmit={(e) =>
        submit(e, async () => {
          const found = validate(form);
          setClientErrors(found);
          if (Object.keys(found).length > 0) {
            setError("");
            // take the user to the first problem once the messages have rendered
            requestAnimationFrame(() =>
              document.querySelector<HTMLElement>('[aria-invalid="true"], [data-invalid="true"]')?.focus(),
            );
            return;
          }
          if (service) {
            await api(`/services/${service.id}`, { method: "PATCH", body: payload(form, service) });
            router.push(`/services/${service.id}`);
          } else {
            const created = await api<Service>("/services", { method: "POST", body: payload(form) });
            // the page is kept when you leave it, so clear it for the next service, token included
            setForm(initialState());
            router.push(`/services/${created.id}`);
          }
        })
      }
      className="space-y-6"
    >
      {formError && <Alert tone="error">{formError}</Alert>}
      {Object.keys(clientErrors).length > 0 && (
        <Alert tone="error">Some fields need attention. Check the messages below.</Alert>
      )}

      <Section title="Endpoint" description="The address Pingwise calls, and how it calls it.">
        <Field
          label="Name"
          name="name"
          value={form.name}
          onChange={(e) => set("name", e.target.value)}
          placeholder="Payments API"
          maxLength={100}
          error={errors.name}
          autoFocus={!service}
        />
        <div className="grid gap-5 sm:grid-cols-[8rem_1fr]">
          <SelectField
            label="Method"
            name="method"
            value={form.method}
            onChange={(e) => set("method", e.target.value as Service["method"])}
            error={errors.method}
          >
            <option>GET</option>
            <option>POST</option>
            <option>HEAD</option>
          </SelectField>
          <Field
            label="URL"
            name="url"
            type="url"
            inputMode="url"
            value={form.url}
            onChange={(e) => set("url", e.target.value)}
            placeholder="https://api.example.com/health"
            error={errors.url}
          />
        </div>
        {form.method === "POST" && (
          <TextAreaField
            label="Request body"
            name="request_body"
            rows={5}
            value={form.requestBody}
            onChange={(e) => set("requestBody", e.target.value)}
            placeholder='{"ping": true}'
            maxLength={10000}
            hint="Sent as-is. Add a Content-Type header below if your endpoint needs one."
            error={errors.request_body}
          />
        )}
      </Section>

      <Section title="Expected response" description="A check passes when the endpoint returns this status code in time.">
        <div className="grid gap-5 sm:grid-cols-2">
          <Field
            label="Expected status code"
            name="expected_status"
            type="number"
            min={100}
            max={599}
            value={form.expectedStatus}
            onChange={(e) => set("expectedStatus", e.target.value)}
            error={errors.expected_status}
          />
          <Field
            label="Timeout (seconds)"
            name="timeout_seconds"
            type="number"
            min={1}
            max={30}
            value={form.timeoutSeconds}
            onChange={(e) => set("timeoutSeconds", e.target.value)}
            hint="Between 1 and 30 seconds."
            error={errors.timeout_seconds}
          />
        </div>
      </Section>

      <Section title="Schedule" description="How often to check, and when to treat failures as an outage.">
        <div className="grid gap-5 sm:grid-cols-2">
          <SelectField
            label="Check interval"
            name="interval_minutes"
            value={form.intervalMinutes}
            onChange={(e) => set("intervalMinutes", e.target.value)}
            error={errors.interval_minutes}
          >
            {INTERVALS.map((m) => (
              <option key={m} value={m}>
                {intervalLabel(m)}
              </option>
            ))}
          </SelectField>
          <SelectField
            label="Mark as down after"
            name="failure_threshold"
            value={form.failureThreshold}
            onChange={(e) => set("failureThreshold", e.target.value)}
            hint="A single failure can be a network blip; 2 or more avoids false alarms."
            error={errors.failure_threshold}
          >
            {[1, 2, 3, 4, 5].map((n) => (
              <option key={n} value={n}>
                {n === 1 ? "1 failed check" : `${n} failed checks in a row`}
              </option>
            ))}
          </SelectField>
        </div>
      </Section>

      <Section
        title="Authentication"
        description="For endpoints that need a token. Use a dedicated read-only token where you can."
      >
        <AuthFields form={form} set={set} errors={errors} service={service} />
      </Section>

      <Section title="Custom headers" description="Optional headers sent with every check. Don't put secrets here; use authentication.">
        {form.headers.length > 0 && (
          <ul className="space-y-2">
            {form.headers.map((row, i) => (
              <li key={row.id} className="flex items-start gap-2">
                <input
                  aria-label={`Header ${i + 1} name`}
                  data-invalid={errors.custom_headers?.includes(`"${row.name}"`) ? true : undefined}
                  value={row.name}
                  onChange={(e) => setHeader(row.id, { name: e.target.value })}
                  placeholder="Accept"
                  maxLength={100}
                  className="h-10 w-2/5 rounded-md border border-line bg-surface px-3 font-mono text-sm focus:outline-2 focus:outline-accent"
                />
                <input
                  aria-label={`Header ${i + 1} value`}
                  value={row.value}
                  onChange={(e) => setHeader(row.id, { value: e.target.value })}
                  placeholder="application/json"
                  maxLength={2000}
                  className="h-10 min-w-0 flex-1 rounded-md border border-line bg-surface px-3 font-mono text-sm focus:outline-2 focus:outline-accent"
                />
                <Button
                  variant="ghost"
                  aria-label={`Remove header ${row.name || i + 1}`}
                  onClick={() => set("headers", form.headers.filter((h) => h.id !== row.id))}
                >
                  Remove
                </Button>
              </li>
            ))}
          </ul>
        )}
        {errors.custom_headers && (
          <p role="alert" className="text-sm text-down-ink">
            {errors.custom_headers}
          </p>
        )}
        {form.headers.length < MAX_HEADERS && (
          <Button size="sm" onClick={() => set("headers", [...form.headers, { id: nextRowId++, name: "", value: "" }])}>
            Add header
          </Button>
        )}
      </Section>

      <div className="flex items-center gap-3 border-t border-line pt-6">
        <Button type="submit" variant="primary" disabled={pending}>
          {pending ? "Saving…" : service ? "Save changes" : "Add service"}
        </Button>
        <ButtonLink href={service ? `/services/${service.id}` : "/dashboard"} variant="ghost">
          Cancel
        </ButtonLink>
        {service && <span className="text-sm text-muted">Saving restarts the count of consecutive failures.</span>}
      </div>
    </form>
  );
}

function AuthFields({
  form,
  set,
  errors,
  service,
}: {
  form: FormState;
  set: <K extends keyof FormState>(key: K, value: FormState[K]) => void;
  errors: Record<string, string>;
  service?: Service;
}) {
  const hasSaved = Boolean(service?.auth_token_masked);

  if (form.authMode === "none" || form.authMode === "remove") {
    return (
      <div className="flex flex-wrap items-center gap-3 text-sm">
        <span className="text-muted">
          {form.authMode === "remove" ? "The saved token will be removed when you save." : "No authentication."}
        </span>
        <Button
          size="sm"
          onClick={() => {
            set("authMode", hasSaved && form.authMode === "remove" ? "keep" : "set");
          }}
        >
          {form.authMode === "remove" ? "Keep saved token" : "Add authentication"}
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      <div className="grid gap-5 sm:grid-cols-[14rem_1fr]">
        <Field
          label="Header name"
          name="auth_header_name"
          value={form.authHeaderName}
          onChange={(e) => set("authHeaderName", e.target.value)}
          maxLength={100}
          error={errors.auth_header_name}
        />
        {form.authMode === "keep" ? (
          <div>
            <span className="block text-sm font-medium">Token</span>
            <div className="mt-1.5 flex h-10 items-center justify-between rounded-md border border-line bg-canvas px-3 font-mono text-sm text-muted">
              {service?.auth_token_masked}
              <span className="font-sans text-xs">Saved and encrypted</span>
            </div>
          </div>
        ) : (
          <Field
            label="Token"
            name="auth_token"
            type="password"
            autoComplete="off"
            value={form.authToken}
            onChange={(e) => set("authToken", e.target.value)}
            placeholder="Bearer eyJhbGciOi…"
            maxLength={4000}
            hint="Sent exactly as entered, so include any prefix such as Bearer. Encrypted when saved and never shown again."
            error={errors.auth_token}
          />
        )}
      </div>
      <div className="flex flex-wrap gap-2">
        {form.authMode === "keep" && (
          <Button size="sm" onClick={() => set("authMode", "set")}>
            Replace token
          </Button>
        )}
        {form.authMode === "set" && hasSaved && (
          <Button size="sm" onClick={() => set("authMode", "keep")}>
            Keep saved token
          </Button>
        )}
        <Button
          size="sm"
          variant="ghost"
          onClick={() => {
            set("authToken", "");
            set("authMode", hasSaved ? "remove" : "none");
          }}
        >
          {hasSaved ? "Remove authentication" : "Don't use authentication"}
        </Button>
      </div>
    </div>
  );
}

function Section({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return (
    <section className="grid gap-x-10 gap-y-4 rounded-xl border border-line bg-surface p-5 sm:p-6 lg:grid-cols-[16rem_1fr]">
      <div>
        <h2 className="font-semibold">{title}</h2>
        <p className="mt-1 text-sm text-muted">{description}</p>
      </div>
      <div className="space-y-5">{children}</div>
    </section>
  );
}
