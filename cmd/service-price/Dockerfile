FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY cmd/service-price cmd/service-price
COPY pkg/ pkg
COPY internal/ internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /bin/service-price ./cmd/service-price

FROM scratch
USER 1000

COPY --from=build /bin/service-price /app/service-price

ENTRYPOINT ["/app/service-price"]
