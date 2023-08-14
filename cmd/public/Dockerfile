FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY cmd/public cmd/public
COPY pkg/ pkg
COPY internal/ internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /bin/public ./cmd/public

FROM scratch
USER 1000

COPY --from=build /bin/public /app/public

ENTRYPOINT ["/app/public"]
