FROM golang:1.20.1-buster AS builder


WORKDIR /go/src/github.com/quarksgroup/wg-mesh


COPY go.mod go.sum ./
# Download all dependencies
RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux GARCH=amd64 go build -a -installsuffix cgo -o /usr/bin/wgmesh .

CMD ["/usr/bin/wgmesh"]
