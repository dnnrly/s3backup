FROM golang

RUN apt-get update -y && \
    apt-get install -y awscli

CMD /bin/bash