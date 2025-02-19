package main

import (
	"context"
	"time"

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

func (w *AggregateWorker) RunHourlyAggregate() {
	for {
		var startTime = time.Now()

		for _, monitorId := range w.monitorIds {
			historicalData, err := w.reader.ReadRawHistorical(context.TODO(), monitorId)
			if err != nil {
				log.Error().Err(err).Msgf("failed to read raw historical data for monitor %s", monitorId)
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

			if (len(lastHourData) == 0) {
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

			err = w.writer.WriteHourly(context.TODO(), MonitorHistorical{
				MonitorID: monitorId,
				Status:    averageStatus,
				Latency:   averageLatency,
				Timestamp: fromTime,
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to write hourly aggregate data")
				continue
			}
		}

		// Calculate the time that we allowed to sleep. We should wake up 10 minutes after the `startTime`
		var allowedSleepTime = startTime.Add(10 * time.Minute).Sub(time.Now())
		if allowedSleepTime > 0 {
			time.Sleep(allowedSleepTime)
		}
	}
}

func (w *AggregateWorker) RunDailyAggregate() {
	for {
		var startTime = time.Now()

		for _, monitorId := range w.monitorIds {
			historicalData, err := w.reader.ReadRawHistorical(context.TODO(), monitorId)
			if err != nil {
				log.Error().Err(err).Msgf("failed to read raw historical data for monitor %s", monitorId)
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

			if (len(lastHourData) == 0) {
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

			err = w.writer.WriteDaily(context.TODO(), MonitorHistorical{
				MonitorID: monitorId,
				Status:    averageStatus,
				Latency:   averageLatency,
				Timestamp: fromTime,
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to write daily aggregate data")
				continue
			}
		}

		// Calculate the time that we allowed to sleep. We should wake up 1 hour after the `startTime`
		var allowedSleepTime = startTime.Add(1 * time.Hour).Sub(time.Now())
		if allowedSleepTime > 0 {
			time.Sleep(allowedSleepTime)
		}
	}
}