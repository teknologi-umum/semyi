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
	historicalWriter *MonitorHistoricalWriter
	historicalReader *MonitorHistoricalReader
	centralBroker    *Broker[MonitorHistorical]

	telegramAlertProvider Alerter
	discordAlertProvider  Alerter
	httpAlertProvider     Alerter
	slackAlertProvider    Alerter
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
		err := m.historicalWriter.Write(ctx, monitorHistorical)
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
		if m.telegramAlertProvider == nil && m.discordAlertProvider == nil && m.httpAlertProvider == nil && m.slackAlertProvider == nil {
			log.Warn().Msg("no alert providers are set")
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

		lastRawHistorical, err := m.historicalReader.ReadRawLatest(ctx, uniqueId)
		if err != nil {
			log.Error().Err(err).Msg("failed to get raw latest historical data")
			sentry.GetHubFromContext(ctx).CaptureException(err)
			return
		}

		if lastRawHistorical.Status != status {
			// Add breadcrumb for alert sending
			sentry.AddBreadcrumb(&sentry.Breadcrumb{
				Category: "alert",
				Message:  fmt.Sprintf("Sending alert for monitor %s", uniqueId),
				Level:    sentry.LevelInfo,
				Data: map[string]interface{}{
					"monitor_id": uniqueId,
					"provider":   response.Monitor.AlertProvider,
					"status":     status,
				},
			})

			switch response.Monitor.AlertProvider {
			case AlertProviderTypeTelegram:
				if m.telegramAlertProvider == nil {
					log.Warn().Msg("telegram alert provider is not set")
					return
				}

				err := m.telegramAlertProvider.Send(ctx, alertMessage)
				if err != nil {
					log.Error().Err(err).Msg("failed to send alert")
					sentry.GetHubFromContext(ctx).CaptureException(err)
				}
			case AlertProviderTypeDiscord:
				if m.discordAlertProvider == nil {
					log.Warn().Msg("discord alert provider is not set")
					return
				}

				err := m.discordAlertProvider.Send(ctx, alertMessage)
				if err != nil {
					log.Error().Err(err).Msg("failed to send alert")
					sentry.GetHubFromContext(ctx).CaptureException(err)
				}
			case AlertProviderTypeHTTP:
				if m.httpAlertProvider == nil {
					log.Warn().Msg("http alert provider is not set")
					return
				}

				err := m.httpAlertProvider.Send(ctx, alertMessage)
				if err != nil {
					log.Error().Err(err).Msg("failed to send alert")
					sentry.GetHubFromContext(ctx).CaptureException(err)
				}
			case AlertProviderTypeSlack:
				if m.slackAlertProvider == nil {
					log.Warn().Msg("slack alert provider is not set")
					return
				}

				err := m.slackAlertProvider.Send(ctx, alertMessage)
				if err != nil {
					log.Error().Err(err).Msg("failed to send alert")
					sentry.GetHubFromContext(ctx).CaptureException(err)
				}
			}
		}

		// TODO: incident writer
	}()

	m.centralBroker.Publish(uniqueId, &BrokerMessage[MonitorHistorical]{
		Header: map[string]string{
			"monitor_id": uniqueId,
			"interval":   "raw",
		},
		Body: monitorHistorical,
	})
}
