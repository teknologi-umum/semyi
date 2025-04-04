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
	centralBroker    *Broker[MonitorHistorical]

	telegramAlertProvider Alerter
	discordAlertProvider  Alerter
	httpAlertProvider     Alerter
	slackAlertProvider    Alerter
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

	monitorHistorical := MonitorHistorical{
		MonitorID: uniqueId,
		Status:    status,
		Latency:   response.RequestDuration,
		Timestamp: response.Timestamp,
	}

	attemptRemaining := 3
	attemptedEntries := 0
	for attemptRemaining > 0 {
		err := m.historicalWriter.Write(context.Background(), monitorHistorical)
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

		lastRawHistorical, err := m.historicalReader.ReadRawLatest(context.Background(), uniqueId)
		if err != nil {
			log.Error().Err(err).Msg("failed to get raw latest historical data")
			return
		}

		if lastRawHistorical.Status != status {
			switch response.Monitor.AlertProvider {
			case AlertProviderTypeTelegram:
				if m.telegramAlertProvider == nil {
					log.Warn().Msg("telegram alert provider is not set")
					return
				}

				err := m.telegramAlertProvider.Send(context.Background(), alertMessage)
				if err != nil {
					log.Error().Err(err).Msg("failed to send alert")
				}
			case AlertProviderTypeDiscord:
				if m.discordAlertProvider == nil {
					log.Warn().Msg("discord alert provider is not set")
					return
				}

				err := m.discordAlertProvider.Send(context.Background(), alertMessage)
				if err != nil {
					log.Error().Err(err).Msg("failed to send alert")
				}
			case AlertProviderTypeHTTP:
				if m.httpAlertProvider == nil {
					log.Warn().Msg("http alert provider is not set")
					return
				}

				err := m.httpAlertProvider.Send(context.Background(), alertMessage)
				if err != nil {
					log.Error().Err(err).Msg("failed to send alert")
				}
			case AlertProviderTypeSlack:
				if m.slackAlertProvider == nil {
					log.Warn().Msg("slack alert provider is not set")
					return
				}

				err := m.slackAlertProvider.Send(context.Background(), alertMessage)
				if err != nil {
					log.Error().Err(err).Msg("failed to send alert")
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
