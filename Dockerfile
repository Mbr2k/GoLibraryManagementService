FROM golang:1.16.3-buster

WORKDIR /go/src
COPY . .

RUN cd /go/src 

CMD ["/bin/bash"]

EXPOSE 8080/tcp

