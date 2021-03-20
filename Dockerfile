FROM golang:latest
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]


# FROM golang:1.14

# WORKDIR /go/src/app
# COPY . .

# RUN go get -d -v ./...
# RUN go install -v ./...
# RUN go build -o main .
# CMD ["app"]