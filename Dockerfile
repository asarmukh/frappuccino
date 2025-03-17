FROM golang:1.22.2

WORKDIR /app

COPY . .

RUN go build -o frappuccino .

EXPOSE 8080

CMD ["./frappuccino"]
# edit CMD command accordingly.?
