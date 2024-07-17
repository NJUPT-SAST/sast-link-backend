# ----------------------------------------------------------------
# 1. Build the application
# ----------------------------------------------------------------
FROM golang:1.22 AS builder

ADD src /sastlink
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
RUN cd /sastlink && CGO_ENABLED=0 go build

# ----------------------------------------------------------------
# 2. Create a minimal image
# ----------------------------------------------------------------
FROM alpine

COPY --from=builder /sastlink /sastlink
COPY --from=builder /sastlink/config /config

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/sastlink/sast-link-backend"]
