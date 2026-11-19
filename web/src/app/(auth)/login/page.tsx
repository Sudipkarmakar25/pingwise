import type { Metadata } from "next";
import Link from "next/link";
import { Suspense } from "react";
import { LoginForm } from "./login-form";

export const metadata: Metadata = { title: "Sign in" };

export default function LoginPage() {
  return (
    <>
      <h1 className="text-2xl font-semibold tracking-tight">Sign in to Pingwise</h1>
      <p className="mt-2 text-sm text-muted">Enter the email and password for your account.</p>
      <div className="mt-8 rounded-xl border border-line bg-surface p-6">
        <Suspense>
          <LoginForm />
        </Suspense>
      </div>
      <p className="mt-6 text-center text-sm text-muted">
        New to Pingwise?{" "}
        <Link href="/signup" className="font-medium text-accent-strong hover:underline">
          Create an account
        </Link>
      </p>
    </>
  );
}
