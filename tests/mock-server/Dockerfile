FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY mock-server-linux-amd64 /mock-server
RUN chmod +x /mock-server
USER 65534
ENTRYPOINT ["/mock-server"]
