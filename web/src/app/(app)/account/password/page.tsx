import type { Metadata } from "next";
import { Suspense } from "react";
import { ChangePasswordForm } from "./change-password-form";

export const metadata: Metadata = { title: "Change password" };

export default function ChangePasswordPage() {
  return (
    <div className="mx-auto max-w-md">
      <Suspense>
        <ChangePasswordForm />
      </Suspense>
    </div>
  );
}
