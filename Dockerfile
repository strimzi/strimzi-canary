FROM centos:7

ARG version
ENV VERSION=${version}

ADD target/main /

CMD ["/main"]