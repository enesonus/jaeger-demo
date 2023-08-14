FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY cmd/service-title cmd/service-title
COPY pkg/ pkg
COPY internal/ internal

RUN go build -o /bin/service-title ./cmd/service-title

FROM scratch
USER 1000

COPY --from=build /bin/service-title /app/service-title

ENTRYPOINT ["/app/service-title"]
