FROM golang:1.10.0

ENV GIT_TERMINAL_PROMPT=1

RUN apt update && apt install -y rsync
RUN apt-get update -qq && apt-get install -qqy \
    apt-transport-https \
    ca-certificates \
    curl \
    lxc \
    iptables


# Install Docker from Docker Inc. repositories.
RUN curl -sSL https://get.docker.com/ | sh


RUN go get -u github.com/golang/dep/cmd/dep
RUN go get -u github.com/onsi/ginkgo/ginkgo

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.9.0/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

RUN curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash 


# Run dep ensure to fill in the cache:
RUN mkdir -p ${GOPATH}/src/github.com/solo-io/ && \
    cd ${GOPATH}/src/github.com/solo-io/ && \
    git clone -b monorepo  https://github.com/solo-io/gloo  && \
    cd gloo && \
        dep ensure -v && \
    cd / && \
    rm -rf ${GOPATH}/src/github.com/solo-io/


CMD ["/bin/bash"]
