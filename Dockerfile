FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w -buildid=" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

ENV GIN_MODE=release

WORKDIR /
COPY --from=build /out/api /api

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/api"]
