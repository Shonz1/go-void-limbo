# syntax=docker/dockerfile:1

# The limbo is a single static binary with no dependencies outside the
# standard library, so the image is that binary on top of a root that holds
# nothing but CA certificates (for checking logins with Mojang) and a
# non-root user to run as.
#
# The build stage runs on the builder's own platform and cross-compiles for
# the target, which is what makes one `docker buildx build --platform
# linux/amd64,linux/arm64` produce both images without emulating either.

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/limbo .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/limbo /limbo

# The working directory is where a world is looked for when WORLD is unset:
# mount one at /data/worlds/Lobby, or anywhere and point WORLD at it.
WORKDIR /data

EXPOSE 25565

ENTRYPOINT ["/limbo"]
