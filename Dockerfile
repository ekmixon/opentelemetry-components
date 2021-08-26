FROM golang:1.17 as build
WORKDIR /otel
COPY . .
RUN make build

FROM debian:stretch-slim
WORKDIR /otel
COPY --from=build /otel/build/otelcompcol_linux_amd64 otelcompcol
CMD [ "/otel/otelcompcol" ]