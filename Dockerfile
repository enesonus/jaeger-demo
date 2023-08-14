ARG port=8080
ARG app=service-main

FROM golang:1.19 as builder
ARG app
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY main.go ${app}/
COPY lib/ lib

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /out/${app} ./${app}

# final stage
FROM scratch
ARG app
ARG port
COPY --from=builder /out/${app} /app/

EXPOSE ${port}
ENTRYPOINT ["/app/service-main"]
