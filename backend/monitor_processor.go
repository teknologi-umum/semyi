package main

import "github.com/rs/zerolog/log"

type Processor struct {
	historicalWriter *MonitorHistoricalWriter
	historicalReader *MonitorHistoricalReader
}

func (m *Processor) ProcessResponse(response Response) {
	status := MonitorStatusFailure
	if response.Success {
		status = MonitorStatusSuccess
	}

	uniqueId := response.Monitor.UniqueID
	if len(uniqueId) >= 255 {
		// Truncate the unique ID if it's too long
		uniqueId = uniqueId[:255]
	}

	// TODO: Retry write if it fails
	// Write the response to the historical writer
	err := m.historicalWriter.Write(MonitorHistorical{
		MonitorID: uniqueId,
		Status:    status,
		Latency:   response.RequestDuration,
		Timestamp: response.Timestamp,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to write historical data")
	}

	// TODO: If the current status is different from the last status, send an alert notification
}
