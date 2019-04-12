FROM golang:1.12 as builder
WORKDIR /app/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grafana-datasource-manager .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates && adduser -S gdm
COPY --from=builder /app/grafana-datasource-manager /usr/bin/
USER gdm
CMD ["grafana-datasource-manager"]