# Menreiki
# (c) 2026 Shiranui (https://shiranui.xyz).

FROM golang:alpine AS builder
COPY . /app
WORKDIR /app
RUN apk add --no-cache build-base make
ENV CGO_ENABLED=1
RUN make && make clean-deps

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/build/menreiki .
COPY avatar.png .

ENTRYPOINT ["./menreiki"]
CMD ["serve"]
