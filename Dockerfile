## Build
FROM bitnami/golang:1.19-debian-11 as build

WORKDIR /app

# Download necessary Go modules
COPY . .
RUN git clone https://github.com/vishnubob/wait-for-it.git

# Download necessary Go modules
RUN go build -o /aocryptobot

## Deploy
FROM bitnami/minideb:bullseye as deploy

WORKDIR /
COPY --from=build /aocryptobot /aocryptobot
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/wait-for-it/wait-for-it.sh /wait-for-it.sh

USER 1001:1001

ENTRYPOINT ["/aocryptobot"]



