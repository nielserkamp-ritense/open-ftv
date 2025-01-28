package network

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

func (m *manager) schedule() {
	scheduler, _ := gocron.NewScheduler()

	scheduler.Start()
	m.logger.Info("request scheduler started successfully")

	for i := range m.cfg.Sources {
		source := m.cfg.Sources[i]

		for j := range source.Requests {
			request := source.Requests[j]
			logger := m.logger.With("source", source.Name, "request", request.Name)

			var def gocron.JobDefinition

			switch {
			case request.Interval > 0:
				def = gocron.DurationJob(request.Interval)
			case request.Schedule != "":
				def = gocron.CronJob(request.Schedule, false)
			default:
				logger.Error("failed to schedule job", "error", fmt.Errorf("no interval or schedule defined"))
			}

			if def != nil {
				if _, err := scheduler.NewJob(def, gocron.NewTask(func() { _ = m.execute(logger, request) })); err != nil {
					logger.Error("failed to schedule job", "error", err)
				}
			}
		}
	}

	select {
	case <-m.ctx.Done():
		if err := scheduler.Shutdown(); err != nil {
			m.logger.Info("request scheduler failed to stop", "error", err)
		} else {
			m.logger.Info("request scheduler stopped successfully")
		}
		return
	}
}
