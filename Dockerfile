FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o frappiccino .

EXPOSE 8080

CMD ["./frappiccino"]
# edit CMD command accordingly.?
