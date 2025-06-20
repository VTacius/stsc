package main

import (
	"bytes"
	"fmt"
	"net/http"

	"time"

	"github.com/danopstech/starlink_exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

func main() {
	address := "192.168.100.1:9200"
	intervalo := 5

	duration := time.Duration(intervalo * 1000000000)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	destino := "https://smme.innovacion.gob.sv/api/v1/push"
	for {
		metrics, err := obtenerMetricas(address)
		if err != nil {
			log.Error(err)
		}

		if err := enviarMetrics(destino, metrics); err != nil {
			log.Error(err)
		}

		<-ticker.C
	}
}

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

func enviarMetrics(destino string, data []*io_prometheus_client.MetricFamily) error {
	var metrics bytes.Buffer
	encode := expfmt.NewEncoder(&metrics, expfmt.FmtOpenMetrics)
	for _, metrica := range data {
		encode.Encode(metrica)
	}

	req, err := http.NewRequest("POST", destino, bytes.NewBufferString(metrics.String()))
	if err != nil {
		return err
	}

	// Configurar headers
	req.Header.Set("Content-Type", string(expfmt.FmtText))
	req.Header.Set("X-Scope-OrgID", "Innovación") // Header necesario para Mimir

	// Enviar petición
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("respuesta inesperada: %s", resp.Status)
	}

	return nil
}
