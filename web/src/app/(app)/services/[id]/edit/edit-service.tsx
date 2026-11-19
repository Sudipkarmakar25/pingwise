"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { BackLink } from "@/components/back-link";
import { StatusBadge, displayStatus } from "@/components/status-badge";
import { Button, ButtonLink } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { api, ApiError } from "@/lib/api";
import type { Service } from "@/lib/types";
import { EditAccess } from "../../_components/edit-access";
import { ServiceForm } from "../../_components/service-form";

type Load = { state: "loading" } | { state: "error"; status: number; message: string } | { state: "ready"; service: Service };

export function EditService() {
  const { id } = useParams<{ id: string }>();
  const [load, setLoad] = useState<Load>({ state: "loading" });

  const fetchService = useCallback(async () => {
    try {
      const { service } = await api<{ service: Service }>(`/services/${id}`);
      setLoad({ state: "ready", service });
    } catch (err) {
      const e = err instanceof ApiError ? err : new ApiError(0, "Something went wrong. Try again.");
      setLoad({ state: "error", status: e.status, message: e.message });
    }
  }, [id]);

  useEffect(() => {
    // the request starts here; state is set only when it resolves
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchService();
  }, [fetchService]);

  if (load.state === "loading") {
    return (
      <div className="space-y-3">
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-8 w-64" />
        <Skeleton className="mt-8 h-96 w-full rounded-xl" />
      </div>
    );
  }

  if (load.state === "error") {
    const notFound = load.status === 404 || load.status === 400;
    return (
      <div className="rounded-xl border border-line bg-surface px-6 py-12 text-center">
        <h1 className="font-semibold">{notFound ? "This service doesn't exist" : "Couldn't load this service"}</h1>
        <p className="mx-auto mt-2 max-w-md text-sm text-muted">
          {notFound ? "It may have been deleted, or the link may be out of date." : load.message}
        </p>
        <div className="mt-6 flex justify-center gap-2">
          {!notFound && <Button onClick={fetchService}>Try again</Button>}
          <ButtonLink href="/dashboard">Back to services</ButtonLink>
        </div>
      </div>
    );
  }

  const { service } = load;
  return (
    <>
      <BackLink href={`/services/${service.id}`}>{service.name}</BackLink>
      <div className="mt-3 flex flex-wrap items-center gap-3">
        <h1 className="text-2xl font-semibold tracking-tight">Edit {service.name}</h1>
        <StatusBadge status={displayStatus(service.current_status, service.is_active)} />
      </div>
      <div className="mt-8 space-y-6">
        <EditAccess>
          <ServiceForm service={service} />
          <ServiceControls service={service} onChange={(s) => setLoad({ state: "ready", service: s })} />
        </EditAccess>
      </div>
    </>
  );
}

// ServiceControls pauses, resumes or deletes the service. These apply at once, separately from the form.
function ServiceControls({ service, onChange }: { service: Service; onChange: (service: Service) => void }) {
  const router = useRouter();
  const [toggling, setToggling] = useState(false);
  const [toggleError, setToggleError] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  async function toggleActive() {
    setToggling(true);
    setToggleError(null);
    try {
      const updated = await api<Service>(`/services/${service.id}`, {
        method: "PATCH",
        body: { is_active: !service.is_active },
      });
      onChange(updated);
    } catch (err) {
      setToggleError(err instanceof ApiError ? err.message : "Something went wrong. Try again.");
    } finally {
      setToggling(false);
    }
  }

  async function deleteService() {
    setDeleting(true);
    setDeleteError(null);
    try {
      await api(`/services/${service.id}`, { method: "DELETE" });
      router.push("/dashboard");
    } catch (err) {
      setDeleteError(err instanceof ApiError ? err.message : "Something went wrong. Try again.");
      setDeleting(false);
    }
  }

  return (
    <section className="divide-y divide-line rounded-xl border border-line bg-surface">
      <div className="flex flex-wrap items-center justify-between gap-4 p-5 sm:p-6">
        <div>
          <h2 className="font-semibold">{service.is_active ? "Pause monitoring" : "Resume monitoring"}</h2>
          <p className="mt-1 text-sm text-muted">
            {service.is_active
              ? "Stop checking this service for now. Its history is kept, and you can resume at any time."
              : "Checking is paused. Resume to check this service again within a minute."}
          </p>
          {toggleError && (
            <p role="alert" className="mt-2 text-sm text-down-ink">
              {toggleError}
            </p>
          )}
        </div>
        <Button onClick={toggleActive} disabled={toggling}>
          {toggling ? "Saving…" : service.is_active ? "Pause" : "Resume"}
        </Button>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-4 p-5 sm:p-6">
        <div>
          <h2 className="font-semibold text-down-ink">Delete service</h2>
          <p className="mt-1 text-sm text-muted">Stop checking this service and remove it from your dashboard.</p>
        </div>
        <Button
          variant="danger"
          onClick={() => {
            setDeleteError(null);
            setConfirming(true);
          }}
        >
          Delete
        </Button>
      </div>
      <ConfirmDialog
        open={confirming}
        title={`Delete ${service.name}?`}
        confirmLabel="Delete service"
        pendingLabel="Deleting…"
        pending={deleting}
        error={deleteError}
        onConfirm={deleteService}
        onClose={() => !deleting && setConfirming(false)}
      >
        Pingwise will stop checking <span className="font-medium text-ink">{service.url}</span>, and the service and
        its history will no longer appear in your dashboard. This can&apos;t be undone.
      </ConfirmDialog>
    </section>
  );
}
