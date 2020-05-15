FROM golang
WORKDIR /go/src/app
COPY . .
RUN \
  go get -d -v ./... & \
  go install -v ./... & \
  go test -v ./...

CMD ["go", "run", "./main.go"]