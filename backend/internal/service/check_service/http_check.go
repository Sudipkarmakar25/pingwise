package check_service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"

	"pingwise/internal/models"
)

const (
	maxRedirects     = 5
	maxBodyReadBytes = 1 << 20 // the body is not checked, it is read only so the connection can be reused
	userAgent        = "Pingwise/1.0 (+uptime monitor)"
)

// error values stored in check_results.error
const (
	errStatusMismatch     = "status_mismatch"
	errTimeout            = "timeout"
	errDnsFailure         = "dns_failure"
	errConnectionRefused  = "connection_refused"
	errTlsError           = "tls_error"
	errTooManyRedirects   = "too_many_redirects"
	errRequestFailed      = "request_failed"
	errTokenDecryptFailed = "token_decrypt_failed"
)

var errRedirectLimit = errors.New("too many redirects")

var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return errRedirectLimit
		}
		return nil
	},
}

type checkOutcome struct {
	IsUp           bool
	StatusCode     *int
	ResponseTimeMs int
	Error          *string
}

// performCheck sends the request and compares the status code with the expected one
func performCheck(ctx context.Context, svc *models.Service, authToken string) *checkOutcome {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(svc.TimeoutSeconds)*time.Second)
	defer cancel()

	var body io.Reader
	if svc.RequestBody != nil {
		body = strings.NewReader(*svc.RequestBody)
	}
	req, err := http.NewRequestWithContext(ctx, svc.Method, svc.Url, body)
	if err != nil {
		return failed(errRequestFailed, 0)
	}
	req.Header.Set("User-Agent", userAgent)
	for name, value := range svc.CustomHeaders {
		req.Header.Set(name, value)
	}
	if svc.AuthHeaderName != nil && authToken != "" {
		req.Header.Set(*svc.AuthHeaderName, authToken)
	}

	start := time.Now()
	resp, err := httpClient.Do(req)
	responseTimeMs := int(time.Since(start).Milliseconds())
	if err != nil {
		return failed(classifyError(err), responseTimeMs)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxBodyReadBytes))

	statusCode := resp.StatusCode
	outcome := &checkOutcome{
		IsUp:           statusCode == svc.ExpectedStatus,
		StatusCode:     &statusCode,
		ResponseTimeMs: responseTimeMs,
	}
	if !outcome.IsUp {
		errValue := errStatusMismatch
		outcome.Error = &errValue
	}
	return outcome
}

func failed(errValue string, responseTimeMs int) *checkOutcome {
	return &checkOutcome{IsUp: false, ResponseTimeMs: responseTimeMs, Error: &errValue}
}

func classifyError(err error) string {
	var dnsErr *net.DNSError
	var netErr net.Error
	var certErr *tls.CertificateVerificationError
	var unknownAuthorityErr x509.UnknownAuthorityError
	var hostnameErr x509.HostnameError
	var recordHeaderErr tls.RecordHeaderError

	switch {
	case errors.Is(err, errRedirectLimit):
		return errTooManyRedirects
	case errors.Is(err, context.DeadlineExceeded):
		return errTimeout
	case errors.As(err, &dnsErr):
		return errDnsFailure
	case errors.Is(err, syscall.ECONNREFUSED):
		return errConnectionRefused
	case errors.As(err, &certErr), errors.As(err, &unknownAuthorityErr),
		errors.As(err, &hostnameErr), errors.As(err, &recordHeaderErr):
		return errTlsError
	case errors.As(err, &netErr) && netErr.Timeout():
		return errTimeout
	default:
		return errRequestFailed
	}
}
