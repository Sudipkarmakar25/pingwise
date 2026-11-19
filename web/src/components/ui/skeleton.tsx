// Skeleton is a grey placeholder in the shape of content that is still loading
export function Skeleton({ className = "" }: { className?: string }) {
  return <div aria-hidden="true" className={`animate-pulse rounded-md bg-line/70 ${className}`} />;
}
