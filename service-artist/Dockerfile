FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY service-artist service-artist
COPY pkg/ pkg

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /bin/service-artist ./service-artist

FROM scratch
USER 1000

COPY --from=build /bin/service-artist /app/service-artist

ENTRYPOINT ["/app/service-artist"]
