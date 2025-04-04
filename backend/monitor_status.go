package main

type MonitorStatus uint8

const (
	MonitorStatusSuccess MonitorStatus = iota
	MonitorStatusFailure
	MonitorStatusDegradedPerformance
	MonitorStatusUnderMaintenance
	MonitorStatusLimitedAvailability
)

func (s MonitorStatus) String() string {
	switch s {
	case MonitorStatusSuccess:
		return "Success"
	case MonitorStatusFailure:
		return "Failure"
	case MonitorStatusDegradedPerformance:
		return "Degraded Performance"
	case MonitorStatusUnderMaintenance:
		return "Under Maintenance"
	case MonitorStatusLimitedAvailability:
		return "Limited Availability"
	default:
		return "Unknown"
	}
}
