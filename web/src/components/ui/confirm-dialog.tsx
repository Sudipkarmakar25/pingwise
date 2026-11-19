"use client";

import { useEffect, useRef } from "react";
import type { ReactNode } from "react";
import { Button } from "./button";

// ConfirmDialog asks before an action that can't be undone. It uses the native <dialog>, which
// traps focus, closes on Escape and returns focus to the button that opened it.
export function ConfirmDialog({
  open,
  title,
  children,
  confirmLabel,
  pendingLabel,
  pending,
  error,
  onConfirm,
  onClose,
}: {
  open: boolean;
  title: string;
  children: ReactNode;
  confirmLabel: string;
  pendingLabel: string;
  pending: boolean;
  error?: string | null;
  onConfirm: () => void;
  onClose: () => void;
}) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  }, [open]);

  return (
    <dialog
      ref={ref}
      onClose={onClose}
      aria-labelledby="confirm-dialog-title"
      className="m-auto w-[calc(100%-2rem)] max-w-md rounded-xl border border-line bg-surface p-0 text-ink shadow-xl backdrop:bg-hero/50"
    >
      <div className="p-6">
        <h2 id="confirm-dialog-title" className="text-lg font-semibold">
          {title}
        </h2>
        <div className="mt-2 text-sm leading-relaxed text-muted">{children}</div>
        {error && (
          <p role="alert" className="mt-4 text-sm text-down-ink">
            {error}
          </p>
        )}
        <div className="mt-6 flex justify-end gap-2">
          <Button onClick={onClose} disabled={pending} autoFocus>
            Cancel
          </Button>
          <Button variant="danger" onClick={onConfirm} disabled={pending}>
            {pending ? pendingLabel : confirmLabel}
          </Button>
        </div>
      </div>
    </dialog>
  );
}
