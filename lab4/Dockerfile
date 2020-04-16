FROM golang

RUN mkdir -p /go/src/github.com/dat520-2020/assignments/

COPY ./ /go/src/github.com/dat520-2020/assignments/

RUN go install -v github.com/assignments/lab4/app/main.go

ENTRYPOINT /go/bin/app/main.go