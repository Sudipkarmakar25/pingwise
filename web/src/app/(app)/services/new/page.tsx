import type { Metadata } from "next";
import { BackLink } from "@/components/back-link";
import { EditAccess } from "../_components/edit-access";
import { ServiceForm } from "../_components/service-form";

export const metadata: Metadata = { title: "Add service" };

export default function NewServicePage() {
  return (
    <div className="mx-auto max-w-4xl">
      <BackLink href="/dashboard">Services</BackLink>
      <h1 className="mt-3 text-2xl font-semibold tracking-tight">Add service</h1>
      <p className="mt-1 text-sm text-muted">Pingwise starts checking a new service within a minute of adding it.</p>
      <div className="mt-8">
        <EditAccess>
          <ServiceForm />
        </EditAccess>
      </div>
    </div>
  );
}
