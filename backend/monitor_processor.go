package main

import (
	"context"
	"math"
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

	attemptRemaining := 3
	attemptedEntries := 0
	for attemptRemaining > 0 {
		err := m.historicalWriter.Write(context.Background(), MonitorHistorical{
			MonitorID: uniqueId,
			Status:    status,
			Latency:   response.RequestDuration,
			Timestamp: response.Timestamp,
		})
		if err != nil {
			attemptedEntries++
			if attemptRemaining == 0 {
				log.Error().Err(err).Msgf("failed to write historical data. Attempt %d failed\n", attemptedEntries)
				return
			}

			delay := time.Second * time.Duration(math.Pow(2, math.Abs(float64(attemptedEntries))))
			log.Error().Err(err).Msgf("failed to write historical data. Attempt %d failed. Retrying in %v...\n", attemptedEntries, delay)

			time.Sleep(delay)

			attemptRemaining -= 1
			continue
		}

		break
	}

	// TODO: If the current status is different from the last status, send an alert notification
}
