# = BUILD: build-env =======================
FROM golang:1.14.2 
# AS build-env
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .

RUN \
  go mod download && \
  apt-get -y update && \
  apt-get install --no-install-recommends --no-install-suggests -y \
    vim 

COPY . .
RUN \
  # go get -d -v ./... && \
  # go install -v ./... && \
  go test -cover -v ./...
  # CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/k8s-exporter-mgr
CMD ["go","run","main.go"]

# = TARGET: app =======================
# FROM golang:1.14.2
# WORKDIR /go/bin/ 
# COPY --from=build-env /go/bin/k8s-exporter-mgr . 
