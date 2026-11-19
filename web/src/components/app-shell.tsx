"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { createContext, Suspense, use, useCallback, useEffect, useRef, useState } from "react";
import type { ReactNode } from "react";
import { api, ApiError } from "@/lib/api";
import type { Me } from "@/lib/types";
import { Brand } from "./brand";
import { Button } from "./ui/button";

// me is null until the signed-in user has loaded
type MeContextValue = { me: Me | null; reloadMe: () => Promise<void> };

const MeContext = createContext<MeContextValue | null>(null);

// useMe returns the signed-in user and organisation for pages inside the app shell
export function useMe() {
  const value = use(MeContext);
  if (!value) throw new Error("useMe must be used inside AppShell");
  return value;
}

const roleLabel = { owner: "Owner", editor: "Editor", viewer: "Viewer" } as const;

// AppShell loads the signed-in user once, draws the top bar, and sends users with a
// temporary password to the change password page before anything else. Pages always
// render, so their static parts appear at once; parts that need the user wait for it.
export function AppShell({ children }: { children: ReactNode }) {
  const [me, setMe] = useState<Me | null>(null);
  const [error, setError] = useState<string | null>(null);

  const reloadMe = useCallback(async () => {
    setError(null);
    try {
      setMe(await api<Me>("/users/me"));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong. Try again.");
    }
  }, []);

  useEffect(() => {
    // the request starts here; state is set only when it resolves
    // eslint-disable-next-line react-hooks/set-state-in-effect
    reloadMe();
  }, [reloadMe]);

  return (
    <div className="flex min-h-screen flex-col">
      {/* reading the URL needs a Suspense boundary on dynamic routes */}
      <Suspense>{me?.user.must_change_password && <PasswordGate />}</Suspense>
      <header className="bg-hero">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between gap-6 px-6">
          <div className="flex items-center gap-4 sm:gap-8">
            <Brand onDark />
            {me && (
              <Suspense fallback={<NavLinks me={me} pathname={null} />}>
                <ActiveNavLinks me={me} />
              </Suspense>
            )}
          </div>
          {me ? <UserMenu me={me} /> : <div className="h-8 w-40 animate-pulse rounded-md bg-white/10" />}
        </div>
      </header>

      <main className="mx-auto w-full max-w-6xl flex-1 px-6 pt-6 pb-10">
        {error ? (
          <div className="mx-auto max-w-md rounded-xl border border-line bg-surface p-6 text-center">
            <h1 className="font-semibold">Pingwise couldn&apos;t load your account</h1>
            <p className="mt-2 text-sm text-muted">{error}</p>
            <Button className="mt-5" onClick={reloadMe}>
              Try again
            </Button>
          </div>
        ) : (
          <MeContext value={{ me, reloadMe }}>{children}</MeContext>
        )}
      </main>
    </div>
  );
}

// PasswordGate sends a user who still has a temporary password to the change password page
function PasswordGate() {
  const router = useRouter();
  const pathname = usePathname();
  useEffect(() => {
    if (pathname !== "/account/password") {
      router.replace(`/account/password?next=${encodeURIComponent(pathname)}`);
    }
  }, [pathname, router]);
  return null;
}

function ActiveNavLinks({ me }: { me: Me }) {
  return <NavLinks me={me} pathname={usePathname()} />;
}

function NavLinks({ me, pathname }: { me: Me; pathname: string | null }) {
  const nav = [
    { href: "/dashboard", label: "Services", show: true },
    { href: "/users", label: "Team", show: me.user.role === "owner" },
  ];
  return (
    <nav aria-label="Main" className="flex items-center gap-1">
      {nav
        .filter((item) => item.show)
        .map((item) => {
          // a service page belongs to Services
          const active =
            pathname !== null &&
            (pathname === item.href ||
              pathname.startsWith(`${item.href}/`) ||
              (item.href === "/dashboard" && pathname.startsWith("/services")));
          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={active ? "page" : undefined}
              className={`rounded-md px-3 py-1.5 text-sm transition-colors ${
                active ? "bg-white/10 text-hero-ink" : "text-hero-muted hover:text-hero-ink"
              }`}
            >
              {item.label}
            </Link>
          );
        })}
    </nav>
  );
}

function UserMenu({ me }: { me: Me }) {
  const [open, setOpen] = useState(false);
  const [signingOut, setSigningOut] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const close = (event: MouseEvent | KeyboardEvent) => {
      if (event instanceof KeyboardEvent ? event.key === "Escape" : !ref.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", close);
    document.addEventListener("keydown", close);
    return () => {
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", close);
    };
  }, [open]);

  // A full page load, not a client navigation: Next.js keeps the state of pages you leave,
  // and none of this user's data may survive into the next session on the same browser.
  async function signOut() {
    setSigningOut(true);
    await api("/auth/logout", { method: "POST" }).catch(() => undefined);
    // eslint-disable-next-line @next/next/no-location-assign-relative-destination
    window.location.assign("/login");
  }

  const initials = me.user.name
    .split(/\s+/)
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={open}
        className="flex items-center gap-3 rounded-md py-1 pr-1 pl-2 text-left hover:bg-white/10"
      >
        <span className="hidden text-right sm:block">
          <span className="block text-sm leading-tight text-hero-ink">{me.organisation.name}</span>
          <span className="block text-xs leading-tight text-hero-muted">{roleLabel[me.user.role]}</span>
        </span>
        <span className="flex h-8 w-8 items-center justify-center rounded-full bg-accent text-xs font-semibold text-accent-ink">
          {initials}
        </span>
      </button>
      {open && (
        <div
          role="menu"
          className="absolute right-0 z-20 mt-2 w-64 rounded-lg border border-line bg-surface py-1.5 text-sm shadow-lg"
        >
          <div className="border-b border-line px-3.5 pt-1.5 pb-2.5">
            <div className="font-medium">{me.user.name}</div>
            <div className="truncate text-muted">{me.user.email}</div>
          </div>
          <Link
            role="menuitem"
            href="/account/password"
            onClick={() => setOpen(false)}
            className="block px-3.5 py-2 hover:bg-canvas"
          >
            Change password
          </Link>
          <button
            role="menuitem"
            type="button"
            onClick={signOut}
            disabled={signingOut}
            className="block w-full px-3.5 py-2 text-left hover:bg-canvas disabled:opacity-60"
          >
            {signingOut ? "Signing out…" : "Sign out"}
          </button>
        </div>
      )}
    </div>
  );
}
