"use client";

import { useCallback, useEffect, useState } from "react";
import { useMe } from "@/components/app-shell";
import { Alert } from "@/components/ui/alert";
import { Button, ButtonLink } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Field, SelectField } from "@/components/ui/field";
import { Skeleton } from "@/components/ui/skeleton";
import { api, ApiError } from "@/lib/api";
import { exactTime } from "@/lib/format";
import type { Role, User } from "@/lib/types";
import { useFormSubmit } from "@/lib/use-form-submit";

const roles: { value: Role; label: string; description: string }[] = [
  { value: "owner", label: "Owner", description: "Everything editors can do, plus managing the team." },
  { value: "editor", label: "Editor", description: "Add, change, pause and delete services." },
  { value: "viewer", label: "Viewer", description: "See services, status and history." },
];
const roleLabel = Object.fromEntries(roles.map((r) => [r.value, r.label])) as Record<Role, string>;

// Backend messages that need a plainer explanation
function friendlyError(err: unknown) {
  if (!(err instanceof ApiError)) return "Something went wrong. Try again.";
  if (err.message.includes("at least one owner")) {
    return "Your organisation needs at least one owner. Make someone else an owner first.";
  }
  return err.message;
}

type Created = { name: string; email: string; password: string };

export function TeamView() {
  const { me } = useMe();
  const [users, setUsers] = useState<User[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [adding, setAdding] = useState(false);
  const [created, setCreated] = useState<Created | null>(null);
  const isOwner = me?.user.role === "owner";

  const load = useCallback(async () => {
    try {
      setUsers(await api<User[]>("/users"));
      setError(null);
    } catch (err) {
      setError(friendlyError(err));
    }
  }, []);

  useEffect(() => {
    if (!isOwner) return;
    // the request starts here; state is set only when it resolves
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [isOwner, load]);

  if (me && !isOwner) {
    return (
      <div className="mx-auto max-w-xl rounded-xl border border-line bg-surface px-6 py-12 text-center">
        <h1 className="font-semibold">Only owners can manage the team</h1>
        <p className="mx-auto mt-2 max-w-md text-sm text-muted">
          Ask an owner of {me.organisation.name} if someone needs to be added or their access changed.
        </p>
        <ButtonLink href="/dashboard" className="mt-6">
          Back to services
        </ButtonLink>
      </div>
    );
  }

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Team</h1>
          <p className="mt-1 text-sm text-muted">
            People who can sign in to {me?.organisation.name ?? "your organisation"} and what they can do.
          </p>
        </div>
        {!adding && (
          <Button
            variant="primary"
            onClick={() => {
              setCreated(null);
              setAdding(true);
            }}
          >
            Add member
          </Button>
        )}
      </div>

      {created && <CreatedNotice created={created} onDismiss={() => setCreated(null)} />}

      {adding && (
        <AddMemberForm
          onCancel={() => setAdding(false)}
          onAdded={(member) => {
            setAdding(false);
            setCreated(member);
            load();
          }}
        />
      )}

      {error && (
        <div role="alert" className="mt-6 flex items-center justify-between gap-4 rounded-lg border border-down/30 bg-down-soft px-4 py-3 text-sm text-down-ink">
          <span>Couldn&apos;t load the team. {error}</span>
          <Button size="sm" onClick={load}>
            Try again
          </Button>
        </div>
      )}

      <div className="mt-6">
        {users && me ? (
          <MemberList users={users} meId={me.user.id} onChanged={load} />
        ) : (
          !error && <Skeleton className="h-64 rounded-xl" />
        )}
      </div>

      <dl className="mt-4 grid gap-3 sm:grid-cols-3">
        {roles.map((r) => (
          <div key={r.value} className="rounded-lg border border-line bg-surface px-4 py-3 text-sm">
            <dt>
              <RoleBadge role={r.value} />
            </dt>
            <dd className="mt-1.5 text-muted">{r.description}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}

function MemberList({ users, meId, onChanged }: { users: User[]; meId: string; onChanged: () => void }) {
  const [rowError, setRowError] = useState<string | null>(null);
  const [savingId, setSavingId] = useState<string | null>(null);
  const [removing, setRemoving] = useState<User | null>(null);
  const [removePending, setRemovePending] = useState(false);
  const [removeError, setRemoveError] = useState<string | null>(null);
  const names = new Map(users.map((u) => [u.id, u.name]));

  async function changeRole(user: User, role: Role) {
    setSavingId(user.id);
    setRowError(null);
    try {
      await api(`/users/${user.id}`, { method: "PATCH", body: { role } });
      onChanged();
    } catch (err) {
      setRowError(`Couldn't change ${user.name}'s role. ${friendlyError(err)}`);
    } finally {
      setSavingId(null);
    }
  }

  async function remove() {
    if (!removing) return;
    setRemovePending(true);
    setRemoveError(null);
    try {
      await api(`/users/${removing.id}`, { method: "DELETE" });
      setRemoving(null);
      onChanged();
    } catch (err) {
      setRemoveError(friendlyError(err));
    } finally {
      setRemovePending(false);
    }
  }

  return (
    <section className="rounded-xl border border-line bg-surface">
      <div className="flex items-center justify-between border-b border-line px-5 py-4">
        <h2 className="font-semibold">Members</h2>
        <span className="text-sm text-muted">
          {users.length} {users.length === 1 ? "person" : "people"}
        </span>
      </div>
      {rowError && (
        <div className="px-5 pt-4">
          <Alert tone="error">{rowError}</Alert>
        </div>
      )}
      <div className="overflow-x-auto">
        <table className="w-full min-w-[44rem] text-sm">
          <thead>
            <tr className="border-b border-line bg-canvas/60 text-left text-xs font-medium text-muted">
              <th scope="col" className="px-5 py-2.5 font-medium">Member</th>
              <th scope="col" className="w-36 px-3 py-2.5 font-medium">Role</th>
              <th scope="col" className="w-40 px-3 py-2.5 font-medium">Status</th>
              <th scope="col" className="w-44 px-3 py-2.5 font-medium">Added</th>
              <th scope="col" className="w-24 px-5 py-2.5">
                <span className="sr-only">Actions</span>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-line">
            {users.map((user) => {
              const isMe = user.id === meId;
              const addedBy = user.added_by ? names.get(user.added_by) : null;
              return (
                <tr key={user.id} className="align-middle hover:bg-canvas/50">
                  <td className="px-5 py-3">
                    <div className="flex items-center gap-3">
                      <Initials name={user.name} />
                      <div className="min-w-0">
                        <div className="flex items-center gap-2 font-medium">
                          <span className="truncate">{user.name}</span>
                          {isMe && (
                            <span className="rounded-full bg-subtle px-1.5 py-px text-[11px] font-normal text-muted">You</span>
                          )}
                        </div>
                        <div className="truncate text-muted">{user.email}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-3 py-3">
                    {isMe ? (
                      <RoleBadge role={user.role} />
                    ) : (
                      <>
                        <label className="sr-only" htmlFor={`role-${user.id}`}>
                          Role for {user.name}
                        </label>
                        <select
                          id={`role-${user.id}`}
                          value={user.role}
                          disabled={savingId === user.id}
                          onChange={(e) => changeRole(user, e.target.value as Role)}
                          className="h-8 w-28 rounded-md border border-line bg-surface px-2 text-sm disabled:opacity-60"
                        >
                          {roles.map((r) => (
                            <option key={r.value} value={r.value}>
                              {r.label}
                            </option>
                          ))}
                        </select>
                      </>
                    )}
                  </td>
                  <td className="px-3 py-3">
                    {user.must_change_password ? (
                      <span
                        className="inline-flex items-center gap-1.5 text-partial-ink"
                        title="They'll choose their own password when they first sign in"
                      >
                        <span className="h-1.5 w-1.5 rounded-full bg-partial" aria-hidden="true" />
                        Invited
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 text-up-ink">
                        <span className="h-1.5 w-1.5 rounded-full bg-up" aria-hidden="true" />
                        Active
                      </span>
                    )}
                  </td>
                  <td className="px-3 py-3 text-muted">
                    <div title={exactTime(user.created_at)}>
                      {new Date(user.created_at).toLocaleDateString("en-GB", { dateStyle: "medium" })}
                    </div>
                    {addedBy && <div className="truncate text-xs">by {addedBy}</div>}
                  </td>
                  <td className="px-5 py-3 text-right">
                    {!isMe && (
                      <Button
                        size="sm"
                        variant="muted"
                        onClick={() => {
                          setRemoveError(null);
                          setRemoving(user);
                        }}
                        aria-label={`Remove ${user.name}`}
                      >
                        Remove
                      </Button>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      <ConfirmDialog
        open={removing !== null}
        title={`Remove ${removing?.name ?? ""}?`}
        confirmLabel="Remove member"
        pendingLabel="Removing…"
        pending={removePending}
        error={removeError}
        onConfirm={remove}
        onClose={() => !removePending && setRemoving(null)}
      >
        <span className="font-medium text-ink">{removing?.email}</span> will be signed out everywhere and won&apos;t be
        able to sign in to this organisation. You can add them again later.
      </ConfirmDialog>
    </section>
  );
}

const roleStyles: Record<Role, string> = {
  owner: "bg-hero text-hero-ink",
  editor: "bg-accent-soft text-accent-strong",
  viewer: "bg-subtle text-ink",
};

function RoleBadge({ role }: { role: Role }) {
  return (
    <span className={`inline-flex rounded-md px-2 py-0.5 text-xs font-medium ${roleStyles[role]}`}>{roleLabel[role]}</span>
  );
}

function Initials({ name }: { name: string }) {
  const initials = name
    .split(/\s+/)
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
  return (
    <span
      aria-hidden="true"
      className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-subtle text-xs font-semibold text-ink"
    >
      {initials}
    </span>
  );
}

const PASSWORD_ALPHABET = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789";

// generatePassword makes a random temporary password without look-alike characters (0/O, 1/l/I)
function generatePassword(length = 14) {
  const values = crypto.getRandomValues(new Uint32Array(length));
  return Array.from(values, (v) => PASSWORD_ALPHABET[v % PASSWORD_ALPHABET.length]).join("");
}

function AddMemberForm({ onCancel, onAdded }: { onCancel: () => void; onAdded: (member: Created) => void }) {
  const [password, setPassword] = useState(generatePassword);
  const [emailTaken, setEmailTaken] = useState(false);
  const { pending, formError, fieldErrors, submit } = useFormSubmit({
    name: "Enter their name, up to 100 characters.",
    email: "Enter a valid email address.",
    role: "Choose a role.",
    temporary_password: "Use between 8 and 72 characters.",
  });

  return (
    <form
      className="mt-6 rounded-xl border border-line bg-surface p-5 sm:p-6"
      onSubmit={(e) =>
        submit(e, async (values) => {
          setEmailTaken(false);
          const member = {
            name: values.name.trim(),
            email: values.email.trim(),
            role: values.role,
            temporary_password: password,
          };
          try {
            await api("/users", { method: "POST", body: member });
          } catch (err) {
            if (err instanceof ApiError && err.status === 409) setEmailTaken(true);
            throw err;
          }
          onAdded({ name: member.name, email: member.email, password });
        })
      }
    >
      <h2 className="font-semibold">Add a member</h2>
      <p className="mt-1 text-sm text-muted">
        Pingwise doesn&apos;t send invitation emails yet. You&apos;ll share the sign-in details yourself; they choose
        their own password the first time they sign in.
      </p>
      <div className="mt-5 space-y-5">
        {emailTaken ? (
          <Alert tone="error">This email already has a Pingwise account. Each person can belong to one organisation.</Alert>
        ) : (
          formError && <Alert tone="error">{formError}</Alert>
        )}
        <div className="grid gap-5 sm:grid-cols-2">
          <Field label="Name" name="name" required maxLength={100} autoFocus error={fieldErrors.name} />
          <Field
            label="Email"
            name="email"
            type="email"
            required
            placeholder="name@company.com"
            error={fieldErrors.email}
          />
        </div>
        <div className="grid gap-5 sm:grid-cols-2">
          <SelectField label="Role" name="role" defaultValue="viewer" error={fieldErrors.role}>
            {roles.map((r) => (
              <option key={r.value} value={r.value}>
                {r.label} — {r.description}
              </option>
            ))}
          </SelectField>
          <div>
            <Field
              label="Temporary password"
              name="temporary_password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              maxLength={72}
              autoComplete="off"
              className="[&_input]:font-mono"
              error={fieldErrors.temporary_password}
              hint="Generated for you. They'll replace it at first sign-in."
            />
            <Button size="sm" variant="ghost" className="mt-1 -ml-3" onClick={() => setPassword(generatePassword())}>
              Generate another
            </Button>
          </div>
        </div>
      </div>
      <div className="mt-6 flex gap-3">
        <Button type="submit" variant="primary" disabled={pending}>
          {pending ? "Adding…" : "Add member"}
        </Button>
        <Button variant="ghost" onClick={onCancel} disabled={pending}>
          Cancel
        </Button>
      </div>
    </form>
  );
}

// CreatedNotice shows the new member's sign-in details once, so the owner can pass them on
function CreatedNotice({ created, onDismiss }: { created: Created; onDismiss: () => void }) {
  const [copied, setCopied] = useState(false);
  const signInUrl = typeof window === "undefined" ? "/login" : `${window.location.origin}/login`;
  const text = `You've been added to Pingwise.\nSign in at ${signInUrl}\nEmail: ${created.email}\nTemporary password: ${created.password}\nYou'll be asked to choose your own password when you first sign in.`;

  async function copy() {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  }

  return (
    <div role="status" className="mt-6 rounded-xl border border-up/30 bg-up-soft p-5 text-up-ink">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 className="font-semibold">{created.name} has been added</h2>
          <p className="mt-1 text-sm">
            Share these sign-in details with them. The temporary password won&apos;t be shown again.
          </p>
        </div>
        <div className="flex gap-2">
          <Button size="sm" onClick={copy}>
            {copied ? "Copied" : "Copy details"}
          </Button>
          <Button size="sm" variant="ghost" onClick={onDismiss}>
            Done
          </Button>
        </div>
      </div>
      <dl className="mt-4 grid gap-x-6 gap-y-1 rounded-lg bg-surface/70 p-4 text-sm text-ink sm:grid-cols-[10rem_1fr]">
        <dt className="text-muted">Sign-in page</dt>
        <dd className="font-mono">{signInUrl}</dd>
        <dt className="text-muted">Email</dt>
        <dd className="font-mono">{created.email}</dd>
        <dt className="text-muted">Temporary password</dt>
        <dd className="font-mono">{created.password}</dd>
      </dl>
    </div>
  );
}
