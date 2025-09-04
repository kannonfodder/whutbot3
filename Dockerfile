# Multi-stage build for a small static binary
# Use TARGETARCH build arg (set automatically by docker buildx) to support arm64/amd64 multi-arch builds

FROM golang:1.25-alpine AS build
ARG TARGETARCH
WORKDIR /src

# Cache deps
COPY go.mod go.sum ./
RUN go mod download
RUN echo "TARGETARCH=${TARGETARCH}" && go version && echo "GOARCH used: ${TARGETARCH}"
# Copy and build for the target architecture
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o /whutbot3 main.go

# Final image (alpine multi-arch)
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /whutbot3 /usr/local/bin/whutbot3
USER nobody:nobody
ENTRYPOINT ["/usr/local/bin/whutbot3"]
