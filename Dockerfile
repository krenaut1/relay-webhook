FROM alpine:latest
RUN mkdir /app
RUN mkdir /app/config
WORKDIR /app
ADD ./relay-webhook /app
ADD ./config/* /app/config/
ENTRYPOINT ["/app/relay-webhook"]