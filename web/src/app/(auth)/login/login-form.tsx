"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { api, ApiError } from "@/lib/api";
import { safeNext } from "@/lib/navigation";
import type { User } from "@/lib/types";
import { useFormSubmit } from "@/lib/use-form-submit";

export function LoginForm() {
  const router = useRouter();
  const next = safeNext(useSearchParams().get("next"));
  const { pending, formError, fieldErrors, submit, setError } = useFormSubmit(
    {
      email: "Enter a valid email address.",
      password: "Enter your password.",
    },
    { clearOnSuccess: true },
  );

  return (
    <form
      onSubmit={(e) =>
        submit(e, async (values) => {
          try {
            const { user } = await api<{ user: User }>("/auth/login", {
              method: "POST",
              body: { ...values, email: values.email.trim() },
            });
            router.replace(
              user.must_change_password ? `/account/password?next=${encodeURIComponent(next)}` : next,
            );
          } catch (err) {
            if (err instanceof ApiError && err.status === 401) {
              setError("That email and password don't match. Check them and try again.");
              return;
            }
            throw err;
          }
        })
      }
      className="space-y-5"
    >
      {formError && <Alert tone="error">{formError}</Alert>}
      <Field
        label="Email"
        name="email"
        type="email"
        autoComplete="email"
        placeholder="name@company.com"
        required
        autoFocus
        error={fieldErrors.email}
      />
      <Field
        label="Password"
        name="password"
        type="password"
        autoComplete="current-password"
        required
        error={fieldErrors.password}
      />
      <Button type="submit" variant="primary" className="w-full" disabled={pending}>
        {pending ? "Signing in…" : "Sign in"}
      </Button>
    </form>
  );
}
