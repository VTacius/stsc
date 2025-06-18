package main

import (
	"bytes"
	"fmt"

	"time"

	"github.com/danopstech/starlink_exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

func obtenerMetricas(destino string) ([]*io_prometheus_client.MetricFamily, error) {
	exporter, err := exporter.New(destino)
	if err != nil {
		return nil, fmt.Errorf("could not start exporter: %s", err.Error())
	}
	defer exporter.Conn.Close()
	log.Infof("dish id: %s", exporter.DishID)

	registro := prometheus.NewRegistry()
	registro.MustRegister(exporter)
	metricas, err := registro.Gather()
	if err != nil {
		return nil, fmt.Errorf("could not gather metrics from dish: %s", err.Error())
	}

	return metricas, nil
}

func enviarMetrics(destino string, metrics []*io_prometheus_client.MetricFamily) error {
	var buffer bytes.Buffer
	encode := expfmt.NewEncoder(&buffer, expfmt.FmtOpenMetrics)
	for _, metrica := range metrics {
		encode.Encode(metrica)
	}

	fmt.Printf("%s\n: %s\n", destino, &buffer)

	return nil
}

func main() {
	address := "192.168.100.1:9200"
	intervalo := 5

	duration := time.Duration(intervalo * 1000000000)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		metrics, err := obtenerMetricas(address)
		if err != nil {
			log.Error(err)
		}

		if err := enviarMetrics(address, metrics); err != nil {
			log.Error(err)
		}

		<-ticker.C
	}
}
