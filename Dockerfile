FROM golang:1.25 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/blog . \
    && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12

COPY --from=build /out/blog /usr/local/bin/blog
COPY --from=build --chown=65532:65532 /out/data /data

USER 65532:65532

ENV PORT=8080
ENV DB_PATH=/data/blog.db

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/blog"]
