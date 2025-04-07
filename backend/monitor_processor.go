package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
)

type Processor struct {
	HistoricalWriter      *MonitorHistoricalWriter
	HistoricalReader      *MonitorHistoricalReader
	CentralBroker         *Broker[MonitorHistorical]
	TelegramAlertProvider Alerter
	DiscordAlertProvider  Alerter
	HTTPAlertProvider     Alerter
	SlackAlertProvider    Alerter
}

func (m *Processor) ProcessResponse(ctx context.Context, response Response) {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("Processor.ProcessResponse"))
	span.SetData("semyi.monitor.id", response.Monitor.UniqueID)
	ctx = span.Context()
	defer span.Finish()

	status := MonitorStatusFailure
	if response.Success {
		status = MonitorStatusSuccess
	}

	uniqueId := response.Monitor.UniqueID
	if len(uniqueId) >= 255 {
		// Truncate the unique ID if it's too long
		uniqueId = uniqueId[:255]
	}

	// Add breadcrumb for monitor processing
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "monitor",
		Message:  fmt.Sprintf("Processing response for monitor %s", uniqueId),
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"monitor_id": uniqueId,
			"status":     status,
			"success":    response.Success,
			"latency":    response.RequestDuration,
		},
	})

	monitorHistorical := MonitorHistorical{
		MonitorID: uniqueId,
		Status:    status,
		Latency:   response.RequestDuration,
		Timestamp: response.Timestamp,
	}

	attemptRemaining := 3
	attemptedEntries := 0
	for attemptRemaining > 0 {
		err := m.HistoricalWriter.Write(ctx, monitorHistorical)
		if err != nil {
			attemptedEntries++
			if attemptRemaining == 0 {
				log.Error().Err(err).Msgf("failed to write historical data. Attempt %d failed\n", attemptedEntries)
				sentry.GetHubFromContext(ctx).CaptureException(err)
				return
			}

			delay := time.Second * time.Duration(math.Pow(2, math.Abs(float64(attemptedEntries))))
			log.Error().Err(err).Msgf("failed to write historical data. Attempt %d failed. Retrying in %v...\n", attemptedEntries, delay)
			sentry.GetHubFromContext(ctx).CaptureException(err)

			time.Sleep(delay)

			attemptRemaining -= 1
			continue
		}

		break
	}

	go func() {
		if m.TelegramAlertProvider == nil && m.DiscordAlertProvider == nil && m.HTTPAlertProvider == nil && m.SlackAlertProvider == nil {
			log.Warn().Msg("no alert providers are set, skipping alert")
			return
		}

		alertMessage := AlertMessage{
			Success:     response.Success,
			MonitorID:   uniqueId,
			MonitorName: response.Monitor.Name,
			StatusCode:  response.StatusCode,
			Timestamp:   response.Timestamp,
			Latency:     response.RequestDuration,
		}

		if m.TelegramAlertProvider != nil {
			err := m.TelegramAlertProvider.Send(ctx, alertMessage)
			if err != nil {
				log.Error().Err(err).Msg("failed to send telegram alert")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}
		}

		if m.DiscordAlertProvider != nil {
			err := m.DiscordAlertProvider.Send(ctx, alertMessage)
			if err != nil {
				log.Error().Err(err).Msg("failed to send discord alert")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}
		}

		if m.HTTPAlertProvider != nil {
			err := m.HTTPAlertProvider.Send(ctx, alertMessage)
			if err != nil {
				log.Error().Err(err).Msg("failed to send http alert")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}
		}

		if m.SlackAlertProvider != nil {
			err := m.SlackAlertProvider.Send(ctx, alertMessage)
			if err != nil {
				log.Error().Err(err).Msg("failed to send slack alert")
				sentry.GetHubFromContext(ctx).CaptureException(err)
			}
		}
	}()

	m.CentralBroker.Publish(uniqueId, &BrokerMessage[MonitorHistorical]{
		Header: map[string]string{
			"monitor_id": uniqueId,
			"interval":   "raw",
		},
		Body: monitorHistorical,
	})
}
