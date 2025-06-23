package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"time"

	"github.com/danopstech/starlink_exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

const INTERVALO = "10"
const STARLINK = "192.168.100.1:9200"
const IDENTIFICADOR = "localidad"

var items = []string{"uplink_throughput_bytes", "valid_seconds", "pop_ping_drop_ratio", "alert_mast_not_near_vertical", "alert_motors_stuck", "alert_slow_eth_speeds", "alert_thermal_shutdown", "alert_thermal_throttle", "alert_unexpected_location"}

func getEnvar(presente string, clave string) (resultado string) {
	if valor, present := os.LookupEnv(clave); present {
		resultado = valor
	} else {
		resultado = presente
	}
	return
}

func getEnvarInterval() (intervalo int) {
	// Empieza el trabajo de verdad
	valor_intervalo := getEnvar(INTERVALO, "INTERVALO")

	if v, e := strconv.Atoi(valor_intervalo); e == nil {
		intervalo = v
	} else {
		intervalo = 10
	}

	return
}

func getEnvarIndentifier() string {
	return strings.ToLower(strings.ReplaceAll(getEnvar(IDENTIFICADOR, "IDENTIFICADOR"), " ", "_"))
}

func getMetrics(destino string) ([]*io_prometheus_client.MetricFamily, error) {
	exporter, err := exporter.New(destino)
	if err != nil {
		return nil, fmt.Errorf("could not start exporter: %s", err.Error())
	}
	defer exporter.Conn.Close()

	registro := prometheus.NewRegistry()
	registro.MustRegister(exporter)
	metricas, err := registro.Gather()
	if err != nil {
		return nil, fmt.Errorf("could not gather metrics from dish: %s", err.Error())
	}

	return metricas, nil
}

func convertToString(metrics []*io_prometheus_client.MetricFamily) (string, error) {
	var buffer bytes.Buffer
	encode := expfmt.NewEncoder(&buffer, expfmt.FmtOpenMetrics)
	for _, metrica := range metrics {
		encode.Encode(metrica)
	}

	return buffer.String(), nil
}

func processHeader(header string) (resultado string) {

	data := strings.Split(header, "_")
	resultado = strings.Join(data[2:], "_")

	return
}

func processLine(line string) (encabezado string, valor string) {

	if strings.HasPrefix(line, "starlink_dish") {
		data := strings.Split(line, " ")
		encabezado = processHeader(data[0])
		valor = data[1]
	}

	return
}

func parseData(data string) map[string]string {
	resultado := make(map[string]string)

	for _, line := range strings.Split(data, "\n") {
		encabezado, valor := processLine(line)
		if slices.Contains(items, encabezado) {
			resultado[encabezado] = valor
		}
	}

	return resultado
}

func crear_peticion(identificador string, datos map[string]string) (resultado string) {
	var valores string
	for key, value := range datos {
		valores += fmt.Sprintf("%s=%s,", key, value)
	}
	valores = strings.Trim(valores, ",")
	resultado = fmt.Sprintf("starlink_dish,hostname=%s %s", identificador, valores)

	return
}

func sendMetrics(url, username, password, data string) error {
	// Crear solicitud HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		return err
	}

	// Configurar autenticación básica
	req.SetBasicAuth(username, password)

	// Configurar headers
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("User-Agent", "stsc/0.3")

	// Enviar solicitud
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("respuesta inesperada: %s", resp.Status)
	}

	return nil
}

func main() {

	starlink := getEnvar(STARLINK, "STARLINK")
	interval := getEnvarInterval()
	identifier := getEnvarIndentifier()

	duration := time.Duration(interval * 1_000_000_000)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		metrics, err := getMetrics(starlink)
		if err != nil {
			log.Error(err)
			continue
		}

		data, err := convertToString(metrics)
		if err != nil {
			log.Error(err)
			continue
		}

		resultado := parseData(data)

		body := crear_peticion(identifier, resultado)
		println(body)

		if err := sendMetrics("https://smme.innovacion.gob.sv/api/v2/write", "victoria", "jVwUARKlQnDh3H9DcaKY", body); err != nil {
			log.Error(err)
			continue
		}

		fmt.Println("Envío satisfactorio")
		<-ticker.C
	}
}
