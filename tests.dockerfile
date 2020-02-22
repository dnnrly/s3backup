FROM golang

RUN apt-get update -y && \
    apt-get install -y awscli jq groff python3-pip

RUN git clone https://github.com/bats-core/bats-core.git /tmp/bats && \
	/tmp/bats/install.sh /usr/local

ENV GO111MODULE=on

RUN go get -v github.com/mfridman/tparse && \
    go get -v github.com/mikefarah/yq/v3

RUN pip3 install boto3

ENV PATH=/go/bin:${PATH}

CMD /bin/bash