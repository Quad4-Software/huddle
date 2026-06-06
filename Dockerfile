# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/node@sha256:6a2f933ba154d90d2a0c175e292242d060d0ce82303c4b9fc27bc296b258d620 AS web-builder

USER root

WORKDIR /app/web

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./

RUN corepack enable \
  && corepack prepare pnpm@11.0.0 --activate \
  && pnpm install --frozen-lockfile

COPY web/ ./

RUN pnpm build

FROM cgr.dev/chainguard/go@sha256:8333b6b6cae0251842b38d5a578d5a5230d73f9a8d9da377c32db53fa0d97b3a AS go-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN rm -rf internal/server/dist/* \
  && mkdir -p internal/server/dist

COPY --from=web-builder /app/web/dist/ ./internal/server/dist/

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/huddle ./cmd/huddle

FROM cgr.dev/chainguard/static@sha256:77d8b8925dc27970ec2f48243f44c7a260d52c49cd778288e4ee97566e0cb75b

COPY --from=go-builder /out/huddle /usr/bin/huddle

EXPOSE 8080
EXPOSE 3478/udp

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/usr/bin/huddle", "healthcheck"]

USER 65532:65532

ENTRYPOINT ["/usr/bin/huddle"]
CMD ["-addr", ":8080"]
