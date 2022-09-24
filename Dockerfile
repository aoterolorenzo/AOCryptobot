## Build
FROM bitnami/golang:1.19-debian-11 as build

WORKDIR /app

RUN echo $GOPATH
RUN ls

# Download necessary Go modules
COPY . .
RUN go build -o /aocryptobot
RUN /aocryptobot

## Deploy
FROM bitnami/minideb:bullseye as deploy

WORKDIR /

COPY --from=build /aocryptobot /aocryptobot

USER nonroot:nonroot

ENTRYPOINT ["/aocryptobot"]



