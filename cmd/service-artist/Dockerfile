FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY cmd/service-artist cmd/service-artist
COPY pkg/ pkg
COPY internal/ internal

RUN go build -o /bin/service-artist ./cmd/service-artist

FROM scratch
USER 1000

COPY --from=build /bin/service-artist /app/service-artist

ENTRYPOINT ["/app/service-artist"]
