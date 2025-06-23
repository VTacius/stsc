FROM alpine:latest as certs
RUN apk add -U --no-cache ca-certificates

FROM scratch 
ENV STARLINK="192.168.100.1:9200"
ENV INTERVALO="5"
ENV IDENTIFICADOR="localidad"
COPY build/stcs /
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/stcs"]
