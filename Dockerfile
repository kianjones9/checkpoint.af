FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod ./
COPY cmd/ cmd/
COPY internal/ internal/
COPY go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o /checkpoint ./cmd/cli
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine
COPY --from=build /checkpoint /usr/bin/checkpoint
COPY --from=build /server /usr/bin/server
ENTRYPOINT ["server"]