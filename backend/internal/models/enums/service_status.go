package enums

type ServiceStatus string

const (
	ServiceStatusUp      ServiceStatus = "up"
	ServiceStatusDown    ServiceStatus = "down"
	ServiceStatusUnknown ServiceStatus = "unknown"
)
