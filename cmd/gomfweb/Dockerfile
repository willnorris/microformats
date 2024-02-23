FROM --platform=$BUILDPLATFORM cgr.dev/chainguard/wolfi-base:latest as build
RUN apk update && apk add build-base git openssh go-1.21

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH="${TARGETARCH}" CGO_ENABLED=0 go build -v ./cmd/gomfweb

FROM cgr.dev/chainguard/static:latest

COPY --from=build /work/gomfweb /gomfweb

WORKDIR /
ENTRYPOINT ["/gomfweb"]
CMD []

EXPOSE 4001
