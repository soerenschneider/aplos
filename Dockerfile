FROM golang:1.23.6 AS builder

WORKDIR /src

COPY ./ ./
RUN CGO_ENABLED=0 go build -ldflags="-w -X 'main.BuildVersion=$(git describe --tags --abbrev=0 || echo dev)' -X 'main.CommitHash=$(git rev-parse HEAD)'" -o /aplos

FROM gcr.io/distroless/base

COPY --from=builder "/aplos" /aplos
USER 65532:65532

ENTRYPOINT ["/aplos"]
