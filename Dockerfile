FROM golang:1.9 as builder
WORKDIR /go/src/github.com/screepsplus/grafana-datasource-manager/
COPY main.go .
RUN go get -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grafana-datasource-manager .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/screepsplus/grafana-datasource-manager/grafana-datasource-manager .
CMD ["./grafana-datasource-manager"]