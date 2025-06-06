package main

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
)

type AggregateWorker struct {
	monitorIds []string
	reader     *MonitorHistoricalReader
	writer     *MonitorHistoricalWriter
}

func NewAggregateWorker(monitorIds []string, reader *MonitorHistoricalReader, writer *MonitorHistoricalWriter) *AggregateWorker {
	return &AggregateWorker{monitorIds, reader, writer}
}

func (w *AggregateWorker) RunHourlyAggregate(ctx context.Context) {
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())

	// Run worker every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Debug().Msg("running hourly aggregate")
			w.hourlyAggregate(ctx)
			log.Debug().Msg("finished running hourly aggregate")
		}
	}
}

func (w *AggregateWorker) hourlyAggregate(ctx context.Context) {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("AggregateWorker.RunHourlyAggregate"), sentry.WithTransactionName("RunHourlyAggregate"))
	ctx = span.Context()
	defer span.Finish()

	for _, monitorId := range w.monitorIds {
		historicalData, err := w.reader.ReadRawHistorical(ctx, monitorId, false)
		if err != nil {
			log.Error().Err(err).Msgf("failed to read raw historical data for monitor %s", monitorId)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			continue
		}

		// Filter out from the last hour
		// If right now is 08:29, we should filter out data from 08:00 - 09:00
		// If right now is 08:01, we should filter out data from 08:00 - 09:00
		// If right now is 09:00, we should filter out data from 09:00 - 10:00
		var now = time.Now()
		var fromTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		var toTime = fromTime.Add(1 * time.Hour)
		var lastHourData []MonitorHistorical
		for _, data := range historicalData {
			if data.Timestamp.Equal(fromTime) || (data.Timestamp.After(fromTime) && data.Timestamp.Before(toTime)) {
				lastHourData = append(lastHourData, data)
			}
		}

		if len(lastHourData) == 0 {
			continue
		}

		// Calculate the average latency and status
		var totalLatency int64
		var totalStatus int64
		for _, data := range lastHourData {
			totalLatency += data.Latency
			totalStatus += int64(data.Status)
		}

		var averageLatency = totalLatency / int64(len(lastHourData))
		var averageStatus = MonitorStatus(totalStatus / int64(len(lastHourData)))
		var additionalMessage, httpProtocol, tlsVersion, tlsCipherName string
		var tlsExpiryDate time.Time
		// Additional Semyi-specific information should be acquired from
		// the latest available entry in the hourly aggregate table
		for i := len(lastHourData) - 1; i >= 0; i-- {
			data := lastHourData[i]
			// Additional message should only be set if the status is not success
			if averageStatus != MonitorStatusSuccess && additionalMessage == "" && data.AdditionalMessage != "" {
				additionalMessage = data.AdditionalMessage
			}

			if httpProtocol == "" && data.HttpProtocol != "" {
				httpProtocol = data.HttpProtocol
			}

			if tlsVersion == "" && data.TLSVersion != "" {
				tlsVersion = data.TLSVersion
			}

			if tlsCipherName == "" && data.TLSCipherName != "" {
				tlsCipherName = data.TLSCipherName
			}

			if tlsExpiryDate.IsZero() && !data.TLSExpiryDate.IsZero() {
				tlsExpiryDate = data.TLSExpiryDate
			}

			// If everything has been set, we can break the loop
			if additionalMessage != "" && httpProtocol != "" &&
				tlsVersion != "" && tlsCipherName != "" &&
				!tlsExpiryDate.IsZero() {
				break
			}
		}

		err = w.writer.WriteHourly(ctx, MonitorHistorical{
			MonitorID:         monitorId,
			Status:            averageStatus,
			Latency:           averageLatency,
			Timestamp:         fromTime,
			AdditionalMessage: additionalMessage,
			HttpProtocol:      httpProtocol,
			TLSVersion:        tlsVersion,
			TLSCipherName:     tlsCipherName,
			TLSExpiryDate:     tlsExpiryDate,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to write hourly aggregate data")
			sentry.GetHubFromContext(ctx).CaptureException(err)
			continue
		}
	}
}

func (w *AggregateWorker) RunDailyAggregate(ctx context.Context) {
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())

	// Run worker every 1 hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Debug().Msg("running daily aggregate")
			w.dailyAggregate(ctx)
			log.Debug().Msg("finished running daily aggregate")
		}
	}
}

func (w *AggregateWorker) dailyAggregate(ctx context.Context) {
	span := sentry.StartSpan(ctx, "function", sentry.WithDescription("AggregateWorker.RunDailyAggregate"), sentry.WithTransactionName("RunDailyAggregate"))
	ctx = span.Context()
	defer span.Finish()

	for _, monitorId := range w.monitorIds {
		historicalData, err := w.reader.ReadRawHistorical(ctx, monitorId, false)
		if err != nil {
			log.Error().Err(err).Msgf("failed to read raw historical data for monitor %s", monitorId)
			sentry.GetHubFromContext(ctx).CaptureException(err)
			continue
		}

		// Filter out for today's data
		var now = time.Now()
		var fromTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		var toTime = fromTime.Add(24 * time.Hour)
		var lastHourData []MonitorHistorical
		for _, data := range historicalData {
			if data.Timestamp.Equal(fromTime) || (data.Timestamp.After(fromTime) && data.Timestamp.Before(toTime)) {
				lastHourData = append(lastHourData, data)
			}
		}

		if len(lastHourData) == 0 {
			continue
		}

		// Calculate the average latency and status
		var totalLatency int64
		var totalStatus int64
		for _, data := range lastHourData {
			totalLatency += data.Latency
			totalStatus += int64(data.Status)
		}

		var averageLatency = totalLatency / int64(len(lastHourData))
		var averageStatus = MonitorStatus(totalStatus / int64(len(lastHourData)))
		var additionalMessage, httpProtocol, tlsVersion, tlsCipherName string
		var tlsExpiryDate time.Time
		// Additional Semyi-specific information should be acquired from
		// the latest available entry in the hourly aggregate table
		for i := len(lastHourData) - 1; i >= 0; i-- {
			data := lastHourData[i]
			// Additional message should only be set if the status is not success
			if averageStatus != MonitorStatusSuccess && additionalMessage == "" && data.AdditionalMessage != "" {
				additionalMessage = data.AdditionalMessage
			}

			if httpProtocol == "" && data.HttpProtocol != "" {
				httpProtocol = data.HttpProtocol
			}

			if tlsVersion == "" && data.TLSVersion != "" {
				tlsVersion = data.TLSVersion
			}

			if tlsCipherName == "" && data.TLSCipherName != "" {
				tlsCipherName = data.TLSCipherName
			}

			if tlsExpiryDate.IsZero() && !data.TLSExpiryDate.IsZero() {
				tlsExpiryDate = data.TLSExpiryDate
			}

			// If everything has been set, we can break the loop
			if additionalMessage != "" && httpProtocol != "" &&
				tlsVersion != "" && tlsCipherName != "" &&
				!tlsExpiryDate.IsZero() {
				break
			}
		}

		err = w.writer.WriteDaily(ctx, MonitorHistorical{
			MonitorID:         monitorId,
			Status:            averageStatus,
			Latency:           averageLatency,
			Timestamp:         fromTime,
			AdditionalMessage: additionalMessage,
			HttpProtocol:      httpProtocol,
			TLSVersion:        tlsVersion,
			TLSCipherName:     tlsCipherName,
			TLSExpiryDate:     tlsExpiryDate,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to write daily aggregate data")
			sentry.GetHubFromContext(ctx).CaptureException(err)
			continue
		}
	}
}
