FROM golang:1.7

ENV GOPATH /root/.go
ENV PROJECT  ${GOPATH}/src/github.com/knabben/chatops/consumer/

RUN mkdir -p ${PROJECT}
RUN mkdir -p ${GOPATH}/bin
ENV PATH=${PATH}:${GOPATH}/bin

ADD . ${PROJECT}
WORKDIR ${PROJECT}

RUN curl https://glide.sh/get | sh
RUN glide install
RUN make compile

CMD ["./main", "consume", "--uri", "amqp://guest:guest@rabbitmq:5672"]
