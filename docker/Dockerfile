# Multi-stage build: compile the CLI in a Go toolchain image, then ship a
# minimal runtime image with the source for `go test` and the prebuilt binary.

FROM golang:1.22-alpine AS build

WORKDIR /src
RUN apk add --no-cache git make

COPY go.mod go.sum* ./
RUN if [ -f go.sum ]; then go mod download; fi

COPY . .
RUN go build -o /out/heleket-webhook-inspect ./cmd/heleket-webhook-inspect && \
    CGO_ENABLED=0 go install honnef.co/go/tools/cmd/staticcheck@latest

# ----------------------------------------------------------------------------

FROM golang:1.22-alpine AS dev

RUN apk add --no-cache git make ca-certificates

WORKDIR /app

COPY --from=build /out/heleket-webhook-inspect /usr/local/bin/heleket-webhook-inspect
COPY --from=build /go/bin/staticcheck /usr/local/bin/staticcheck

ENV PATH="/app/bin:${PATH}"

CMD ["go", "version"]
