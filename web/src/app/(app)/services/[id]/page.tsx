import type { Metadata } from "next";
import { Suspense } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { ServiceDetailView } from "./service-detail";

export const metadata: Metadata = { title: "Service" };

export default function ServicePage() {
  return (
    <Suspense fallback={<Skeleton className="h-96 w-full rounded-xl" />}>
      <ServiceDetailView />
    </Suspense>
  );
}
