FROM golang:1.15-alpine AS golang

COPY --from=wcsiu/tdlib:1.7-alpine /usr/local/include/td /usr/local/include/td
COPY --from=wcsiu/tdlib:1.7-alpine /usr/local/lib/libtd* /usr/local/lib/
COPY --from=wcsiu/tdlib:1.7-alpine /usr/lib/libssl.a /usr/local/lib/libssl.a
COPY --from=wcsiu/tdlib:1.7-alpine /usr/lib/libcrypto.a /usr/local/lib/libcrypto.a
COPY --from=wcsiu/tdlib:1.7-alpine /lib/libz.a /usr/local/lib/libz.a
RUN apk add build-base

WORKDIR /demo

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build --ldflags "-extldflags '-static -L/usr/local/lib -ltdjson_static -ltdjson_private -ltdclient -ltdcore -ltdactor -ltddb -ltdsqlite -ltdnet -ltdutils -ldl -lm -lssl -lcrypto -lstdc++ -lz'" -o /tmp/demo-exe main.go

FROM gcr.io/distroless/base:latest
COPY --from=golang /tmp/demo-exe /demo-runner
EXPOSE 3000
ENTRYPOINT [ "/demo-runner" ]
