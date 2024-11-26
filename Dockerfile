FROM golang:1.23 AS build

RUN useradd -u 10001 dimo

WORKDIR /build
COPY . ./

RUN make tidy
RUN make build

FROM gcr.io/distroless/base AS final

LABEL maintainer="DIMO <hello@dimo.zone>"

USER nonroot:nonroot

COPY --from=build --chown=nonroot:nonroot /build/bin/telemetry-api /

EXPOSE 8080
EXPOSE 8888

ENTRYPOINT ["/telemetry-api"]

