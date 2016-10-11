FROM golang:1.6.0

ENV TIME_ZONE=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone

COPY . /go/src/github.com/asiainfoLDP/datafoundry_recharge

WORKDIR /go/src/github.com/asiainfoLDP/datafoundry_recharge

RUN go build

EXPOSE 8090

CMD ["sh", "-c", "./datafoundry_recharge"]