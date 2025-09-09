package metrics

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "goagent.traceable.org/metrics"

type systemMetrics interface {
	getMemory() float64
	getCPU() float64
}

func InitializeSystemMetrics(metricPrefix string) {
	meterProvider := otel.GetMeterProvider()
	meter := meterProvider.Meter(meterName)
	err := setUpMetricRecorder(meter, metricPrefix)
	if err != nil {
		log.Printf("error initializing metrics, failed to setup metric recorder: %v\n", err)
	}
}

func setUpMetricRecorder(meter metric.Meter, metricPrefix string) error {
	finalizedPrefix := ""
	if len(metricPrefix) > 0 {
		finalizedPrefix = metricPrefix + "."
	}

	log.Printf("Setting up metric recorder with prefix: %s\n", metricPrefix)

	if meter == nil {
		return fmt.Errorf("error while setting up metric recorder: meter is nil")
	}
	cpuSeconds, err := meter.Float64ObservableCounter(fmt.Sprintf("%straceable.agent.cpu.seconds.total", finalizedPrefix), metric.WithDescription("Metric to monitor total CPU seconds"))
	if err != nil {
		return fmt.Errorf("error while setting up cpu seconds metric counter: %v", err)
	}
	memory, err := meter.Float64ObservableGauge(fmt.Sprintf("%straceable.agent.memory", finalizedPrefix), metric.WithDescription("Metric to monitor memory usage"))
	if err != nil {
		return fmt.Errorf("error while setting up memory metric counter: %v", err)
	}

	uptime, err := meter.Float64ObservableGauge(fmt.Sprintf("%straceable.agent.uptime", finalizedPrefix), metric.WithDescription("Metric to monitor agent uptime in seconds"))
	if err != nil {
		return fmt.Errorf("error while setting up uptime metric gauge: %v", err)
	}

	// Track process start time for uptime calculation
	processStart := time.Now()

	// Register the callback function for cpu_seconds, memory, and uptime observable gauges
	_, err = meter.RegisterCallback(
		func(ctx context.Context, result metric.Observer) error {
			sysMetrics, err := newSystemMetrics()
			if err != nil {
				return err
			}

			result.ObserveFloat64(cpuSeconds, sysMetrics.getCPU())
			result.ObserveFloat64(memory, sysMetrics.getMemory())
			uptimeSeconds := time.Since(processStart).Seconds()
			result.ObserveFloat64(uptime, uptimeSeconds)

			return nil
		},
		cpuSeconds, memory, uptime,
	)
	if err != nil {
		log.Fatalf("failed to register callback: %v", err)
		return err
	}
	return nil
}
