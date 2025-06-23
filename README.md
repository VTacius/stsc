# Starlink Exporter

Lo único que cambia del proyecto original es que ahora envía ALGUNAS métricas hacia cualquier cosa que entienda el InfluxDB line protocol 

### Contruyendo imagen
```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o build/stcs cmd/starlink_exporter/main.go
```
