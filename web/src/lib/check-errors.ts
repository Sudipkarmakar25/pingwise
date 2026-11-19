// Plain-language labels for the error codes the worker stores with a failed check
const labels: Record<string, string> = {
  status_mismatch: "Unexpected status code",
  timeout: "Timed out",
  dns_failure: "DNS lookup failed",
  connection_refused: "Connection refused",
  tls_error: "TLS or certificate error",
  too_many_redirects: "Too many redirects",
  request_failed: "Request failed",
  token_decrypt_failed: "Couldn't read the saved token",
};

export function checkErrorLabel(code: string | null) {
  if (!code) return null;
  return labels[code] ?? "Check failed";
}
