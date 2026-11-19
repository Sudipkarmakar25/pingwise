import type { Metadata } from "next";
import Link from "next/link";
import { SignupForm } from "./signup-form";

export const metadata: Metadata = { title: "Create an account" };

export default function SignupPage() {
  return (
    <>
      <h1 className="text-2xl font-semibold tracking-tight">Create your Pingwise account</h1>
      <p className="mt-2 text-sm text-muted">
        This creates an organisation with you as its owner. You can invite your team afterwards.
      </p>
      <div className="mt-8 rounded-xl border border-line bg-surface p-6">
        <SignupForm />
      </div>
      <p className="mt-6 text-center text-sm text-muted">
        Already have an account?{" "}
        <Link href="/login" className="font-medium text-accent-strong hover:underline">
          Sign in
        </Link>
      </p>
    </>
  );
}
