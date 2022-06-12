FROM alpine

RUN apk add --no-cache tzdata
COPY build/short-go-app /bin
ENV TZ="Europe/Moscow"

ENTRYPOINT ["/bin/short-go-app"]