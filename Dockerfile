FROM alpine:3.4

COPY gcloud-cleanup gcloud-cleanup
COPY entrypoint.sh entrypoint.sh

RUN apk add --no-cache ca-certificates && \
    chmod +x gcloud-cleanup

ENTRYPOINT ["./gcloud-cleanup"]
