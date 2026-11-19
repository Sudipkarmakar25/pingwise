"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { api, ApiError } from "@/lib/api";
import { useFormSubmit } from "@/lib/use-form-submit";

export function SignupForm() {
  const router = useRouter();
  const [emailTaken, setEmailTaken] = useState(false);
  const { pending, formError, fieldErrors, submit } = useFormSubmit(
    {
      name: "Enter your name, up to 100 characters.",
      email: "Enter a valid email address.",
      password: "Use between 8 and 72 characters.",
      org_name: "Enter your organisation's name, up to 100 characters.",
    },
    { clearOnSuccess: true },
  );

  return (
    <form
      onSubmit={(e) =>
        submit(e, async (values) => {
          setEmailTaken(false);
          try {
            await api("/auth/signup", { method: "POST", body: { ...values, email: values.email.trim() } });
            router.replace("/dashboard");
          } catch (err) {
            if (err instanceof ApiError && err.status === 409) {
              setEmailTaken(true);
            }
            throw err;
          }
        })
      }
      className="space-y-5"
    >
      {emailTaken ? (
        <Alert tone="error">
          An account with this email already exists.{" "}
          <Link href="/login" className="font-medium underline">
            Sign in instead
          </Link>
          .
        </Alert>
      ) : (
        formError && <Alert tone="error">{formError}</Alert>
      )}
      <Field label="Your name" name="name" autoComplete="name" required maxLength={100} autoFocus error={fieldErrors.name} />
      <Field
        label="Work email"
        name="email"
        type="email"
        autoComplete="email"
        placeholder="name@company.com"
        required
        error={fieldErrors.email}
      />
      <Field
        label="Password"
        name="password"
        type="password"
        autoComplete="new-password"
        required
        minLength={8}
        maxLength={72}
        hint="At least 8 characters."
        error={fieldErrors.password}
      />
      <Field
        label="Organisation name"
        name="org_name"
        autoComplete="organization"
        placeholder="Acme Inc."
        required
        maxLength={100}
        error={fieldErrors.org_name}
      />
      <Button type="submit" variant="primary" className="w-full" disabled={pending}>
        {pending ? "Creating account…" : "Create account"}
      </Button>
    </form>
  );
}
