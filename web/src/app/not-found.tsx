import type { Metadata } from "next";
import { Brand } from "@/components/brand";
import { ButtonLink } from "@/components/ui/button";

export const metadata: Metadata = { title: "Page not found" };

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="mx-auto flex h-16 w-full max-w-6xl items-center px-6">
        <Brand />
      </header>
      <main className="flex flex-1 items-center justify-center px-6">
        <div className="max-w-md text-center">
          <p className="font-mono text-sm text-muted">404</p>
          <h1 className="mt-2 text-2xl font-semibold tracking-tight">This page doesn&apos;t exist</h1>
          <p className="mt-3 text-muted">The link may be out of date, or the address may have a typo.</p>
          <div className="mt-6 flex justify-center gap-3">
            <ButtonLink href="/" variant="primary">
              Go to the home page
            </ButtonLink>
            <ButtonLink href="/dashboard">Open dashboard</ButtonLink>
          </div>
        </div>
      </main>
    </div>
  );
}
