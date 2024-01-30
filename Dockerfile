FROM --platform=$BUILDPLATFORM golang:1.21.6 as builder

ENV GO111MODULE=on
ARG TARGETARCH

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags "-s -w" -o rollouts-plugin-trafficrouter-consul .

FROM alpine:3.19.0

ARG TARGETARCH

USER 999

COPY --from=builder /app/rollouts-plugin-trafficrouter-consul /bin/