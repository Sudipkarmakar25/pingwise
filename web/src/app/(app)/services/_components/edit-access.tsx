"use client";

import type { ReactNode } from "react";
import { useMe } from "@/components/app-shell";
import { ButtonLink } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

// EditAccess shows its children only to owners and editors; viewers get an explanation instead
export function EditAccess({ children }: { children: ReactNode }) {
  const { me } = useMe();
  if (!me) return <Skeleton className="h-96 w-full rounded-xl" />;
  if (me.user.role === "viewer") {
    return (
      <div className="rounded-xl border border-line bg-surface px-6 py-12 text-center">
        <h2 className="font-semibold">You have view-only access</h2>
        <p className="mx-auto mt-2 max-w-md text-sm text-muted">
          Only owners and editors can add or change services. Ask one of them if something needs updating.
        </p>
        <ButtonLink href="/dashboard" className="mt-6">
          Back to services
        </ButtonLink>
      </div>
    );
  }
  return children;
}
