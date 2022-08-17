FROM --platform=$BUILDPLATFORM golang:1.19 AS builder
RUN apt-get update && \
    apt-get install -y gcc-aarch64-linux-gnu gcc-x86-64-linux-gnu && \
    ln -s /usr/bin/aarch64-linux-gnu-gcc /usr/bin/arm64-linux-gnu-gcc  && \
    ln -s /usr/bin/x86_64-linux-gnu-gcc /usr/bin/amd64-linux-gnu-gcc

# 1. Precompile the entire go standard library into the first Docker cache layer: useful for other projects too!
RUN CGO_ENABLED=0 GOOS=linux go install -v -installsuffix cgo -a std

ARG TARGETARCH
ARG BUILDVERSION=0.14.4
ARG BUILDVERSION_PRERELEASE
ARG BUILDVERSION_METADATA
ARG TELEMETRY_WRITE_KEY
WORKDIR /go/src/github.com/infrahq/infra

# get deps first so it's cached
COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH \
    CC=$TARGETARCH-linux-gnu-gcc \
    go build \
    -ldflags '-s -X github.com/infrahq/infra/internal.Version='"$BUILDVERSION"' -X github.com/infrahq/infra/internal.Prerelease='"$BUILDVERSION_PRERELEASE"' -X github.com/infrahq/infra/internal.Metadata='"$BUILDVERSION_METADATA"' -X github.com/infrahq/infra/internal.TelemetryWriteKey='"$TELEMETRY_WRITE_KEY"' -linkmode external -extldflags "-static"' \
    .

FROM alpine
COPY --from=builder /go/src/github.com/infrahq/infra/infra /bin/infra
EXPOSE 80
EXPOSE 443
ARG USER=infra
ARG GROUP=infra
RUN addgroup -g 1000 $GROUP && adduser -u 1000 -DG $GROUP $USER
USER $USER:$GROUP
ENTRYPOINT ["/bin/infra"]
CMD ["server"]
