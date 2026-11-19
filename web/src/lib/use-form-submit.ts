"use client";

import { useState } from "react";
import type { FormEvent } from "react";
import { ApiError } from "./api";

// fieldMessages gives each form field the message to show when the API rejects it, so users
// see "Enter a valid email address" rather than the API's internal validation wording.
// clearOnSuccess empties an uncontrolled form after it succeeds; use it for forms with passwords,
// because Next.js keeps pages you leave and the typed values would stay in the page.
export function useFormSubmit(fieldMessages: Record<string, string>, { clearOnSuccess = false } = {}) {
  const [pending, setPending] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  async function submit(event: FormEvent<HTMLFormElement>, action: (values: Record<string, string>) => Promise<void>) {
    event.preventDefault();
    if (pending) return;
    const form = event.currentTarget;
    const values = Object.fromEntries([...new FormData(form).entries()].map(([k, v]) => [k, String(v)]));

    setPending(true);
    setFormError(null);
    setFieldErrors({});
    try {
      await action(values);
    } catch (err) {
      const apiError = err instanceof ApiError ? err : new ApiError(0, "Something went wrong. Try again.");
      const errors: Record<string, string> = {};
      for (const { field } of apiError.validationErrors) {
        if (fieldMessages[field]) errors[field] = fieldMessages[field];
      }
      setFieldErrors(errors);
      if (Object.keys(errors).length === 0) setFormError(apiError.message);
      setPending(false);
      return;
    }
    // Next.js keeps the state of pages you leave, so a form must not stay "saving" after it
    // succeeds: coming back to the page would show a disabled form
    setPending(false);
    if (clearOnSuccess) form.reset();
  }

  // setError shows a message chosen by the caller, for errors it recognises itself
  function setError(message: string) {
    setFormError(message);
    setPending(false);
  }

  return { pending, formError, fieldErrors, submit, setError };
}
