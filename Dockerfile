# = BUILD: build-env =======================
FROM golang:1.14.2 AS build-env
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN \
  go test -cover -v ./... && \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/k8s-prom-exporter-mgr

# = TARGET: app =======================
FROM alpine
WORKDIR /go/bin/
ADD ./example-configs /etc/k8s-prom-exporter-mgr
COPY --from=build-env /go/bin/k8s-prom-exporter-mgr . 
CMD ["go/bin/k8s-prom-exporter-mgr"]