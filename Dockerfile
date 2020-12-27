FROM golang:1.15 AS golang

COPY --from=wcsiu/tdlib:1.7.0 /usr/local/include/td /usr/local/include/td
COPY --from=wcsiu/tdlib:1.7.0 /usr/local/lib/libtd* /usr/local/lib/
COPY --from=wcsiu/tdlib:1.7.0 /usr/lib/x86_64-linux-gnu/libssl.a /usr/local/lib/libssl.a
COPY --from=wcsiu/tdlib:1.7.0 /usr/lib/x86_64-linux-gnu/libcrypto.a /usr/local/lib/libcrypto.a
COPY --from=wcsiu/tdlib:1.7.0 /usr/lib/x86_64-linux-gnu/libz.a /usr/local/lib/libz.a

WORKDIR /demo

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build --ldflags "-extldflags '-static -L/usr/local/lib -ltdjson_static -ltdjson_private -ltdclient -ltdcore -ltdactor -ltddb -ltdsqlite -ltdnet -ltdutils -ldl -lm -lssl -lcrypto -lstdc++ -lz'" -o /tmp/demo-exe main.go

FROM gcr.io/distroless/base:latest
COPY --from=golang /tmp/demo-exe /demo-runner
ENTRYPOINT [ "/demo-runner" ]
