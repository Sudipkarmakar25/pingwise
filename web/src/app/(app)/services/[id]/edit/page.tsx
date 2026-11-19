import type { Metadata } from "next";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { EditService } from "./edit-service";

export const metadata: Metadata = { title: "Edit service" };

export default function EditServicePage() {
  return (
    <div className="mx-auto max-w-4xl">
      <Suspense fallback={<Skeleton className="h-96 w-full rounded-xl" />}>
        <EditService />
      </Suspense>
    </div>
  );
}
