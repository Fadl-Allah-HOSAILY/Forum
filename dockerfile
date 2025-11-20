FROM golang:1.24.0

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o forum

CMD ["./forum"]