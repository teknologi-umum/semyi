package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

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

	maxAttemps := 3
	attempts := 0
	for {
		attempts++
		err := m.historicalWriter.Write(context.Background(), MonitorHistorical{
			MonitorID: uniqueId,
			Status:    status,
			Latency:   response.RequestDuration,
			Timestamp: response.Timestamp,
		})
		if err != nil {
			if attempts >= maxAttemps {
				log.Error().Err(err).Msgf("failed to write historical data. Attempt %d failed\n", attempts)
				return
			}

			delay := time.Duration(attempts*2) * time.Second
			log.Error().Err(err).Msgf("failed to write historical data. Attempt %d failed. Retrying in %v...\n", attempts, delay)

			time.Sleep(delay)
			continue
		}

		break
	}

	// TODO: If the current status is different from the last status, send an alert notification
}
