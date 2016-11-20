FROM golang:latest

RUN mkdir /app

COPY src /app/

WORKDIR /app

RUN go get github.com/docker/docker/api/types
RUN go get github.com/docker/docker/client
RUN go get github.com/gorilla/mux
RUN go get github.com/rs/cors
RUN go get gopkg.in/dancannon/gorethink.v2

RUN go build -o main .

CMD ["/app/main"]

EXPOSE 7000
