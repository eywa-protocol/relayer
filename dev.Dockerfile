# Please keep up to date with the new-version of Golang docker for builder
FROM golang:1.16.0-stretch

RUN apt update && apt upgrade -y && \
    apt install -y git \
    make openssh-client

WORKDIR /p2p-bridge

RUN curl -fLo install.sh https://raw.githubusercontent.com/cosmtrek/air/master/install.sh \
    && chmod +x install.sh && sh install.sh && cp ./bin/air /bin/air
RUN echo ${TYPE_ADAPTER_ENV}
CMD air