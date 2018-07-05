FROM alpine
#RUN apt-get update && apt-get install -y ca-certificates 
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN addgroup -S app && adduser -S -G app appuser 
# RUN groupadd -g 1500 appuser && \
#    useradd -r -u 1500 -g appuser appuser
RUN mkdir /app
RUN chown appuser /app
USER appuser
COPY --from=builder /build/topicServer /app/
WORKDIR /app
EXPOSE 8088/tcp
CMD ["./topicServer"]