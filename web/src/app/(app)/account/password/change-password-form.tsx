"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { useMe } from "@/components/app-shell";
import { Alert } from "@/components/ui/alert";
import { Button, ButtonLink } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { Skeleton } from "@/components/ui/skeleton";
import { api, ApiError } from "@/lib/api";
import { safeNext } from "@/lib/navigation";
import { useFormSubmit } from "@/lib/use-form-submit";

export function ChangePasswordForm() {
  const router = useRouter();
  const { me, reloadMe } = useMe();
  const next = safeNext(useSearchParams().get("next"));
  const [mismatch, setMismatch] = useState(false);
  const [done, setDone] = useState(false);
  const { pending, formError, fieldErrors, submit, setError } = useFormSubmit(
    {
      old_password: "Enter your current password.",
      new_password: "Use between 8 and 72 characters.",
    },
    { clearOnSuccess: true },
  );

  if (!me) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-8 w-56" />
        <Skeleton className="h-4 w-72" />
        <Skeleton className="mt-8 h-80 w-full rounded-xl" />
      </div>
    );
  }
  const required = me.user.must_change_password;

  if (done) {
    return (
      <>
        <h1 className="text-2xl font-semibold tracking-tight">Password changed</h1>
        <div className="mt-6">
          <Alert tone="success">Your new password is in effect. Use it the next time you sign in.</Alert>
        </div>
        <ButtonLink href="/dashboard" className="mt-6">
          Back to services
        </ButtonLink>
      </>
    );
  }

  return (
    <>
      <h1 className="text-2xl font-semibold tracking-tight">{required ? "Set a new password" : "Change password"}</h1>
      <p className="mt-2 text-sm text-muted">
        {required
          ? "Your account was created with a temporary password. Choose your own password to continue."
          : "Enter your current password, then choose a new one."}
      </p>
      <form
        className="mt-8 space-y-5 rounded-xl border border-line bg-surface p-6"
        onSubmit={(e) =>
          submit(e, async ({ old_password, new_password, confirm_password }) => {
            setMismatch(false);
            if (new_password !== confirm_password) {
              setMismatch(true);
              setError("");
              return;
            }
            try {
              await api("/auth/change_password", { method: "POST", body: { old_password, new_password } });
            } catch (err) {
              if (err instanceof ApiError && err.status === 400 && err.validationErrors.length === 0) {
                setError("Your current password is incorrect.");
                return;
              }
              throw err;
            }
            if (required) {
              await reloadMe();
              router.replace(next);
            } else {
              setDone(true);
            }
          })
        }
      >
        {formError && <Alert tone="error">{formError}</Alert>}
        <Field
          label={required ? "Temporary password" : "Current password"}
          name="old_password"
          type="password"
          autoComplete="current-password"
          required
          autoFocus
          error={fieldErrors.old_password}
        />
        <Field
          label="New password"
          name="new_password"
          type="password"
          autoComplete="new-password"
          required
          minLength={8}
          maxLength={72}
          hint="At least 8 characters."
          error={fieldErrors.new_password}
        />
        <Field
          label="Confirm new password"
          name="confirm_password"
          type="password"
          autoComplete="new-password"
          required
          error={mismatch ? "The passwords don't match." : undefined}
        />
        <div className="flex items-center gap-3">
          <Button type="submit" variant="primary" disabled={pending}>
            {pending ? "Saving…" : required ? "Set password and continue" : "Change password"}
          </Button>
          {!required && (
            <ButtonLink href="/dashboard" variant="ghost">
              Cancel
            </ButtonLink>
          )}
        </div>
      </form>
    </>
  );
}
