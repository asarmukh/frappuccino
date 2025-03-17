FROM golang:1.22.2

WORKDIR /app

COPY . .
COPY wait-for-postgres.sh /usr/local/bin/wait-for-postgres.sh
RUN chmod +x /usr/local/bin/wait-for-postgres.sh

RUN go build -o frappuccino .

EXPOSE 8080

CMD ["./frappuccino"]
