FROM golang:1.23-alpine AS build
WORKDIR /app
ADD . .

RUN go build -o create createtotp.go
RUN go build -o load addauthenapptotp.go
RUN go build -o update totpupdateaccount.go
RUN go build -o genqr totpgenqrcode.go
RUN go build -o otp totpshowcode.go
RUN go build -o remove totpremoveaccount.go

FROM alpine AS runtime
WORKDIR /app
COPY --from=build /app/create /bin/create
COPY --from=build /app/load /bin/load
COPY --from=build /app/genqr /bin/genqr
COPY --from=build /app/otp /bin/otp
COPY --from=build /app/remove /bin/remove
COPY --from=build /app/update /bin/update
