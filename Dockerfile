FROM scratch

ADD cmd/target/strimzi-canary ./

LABEL org.opencontainers.image.source='https://github.com/strimzi/strimzi-canary'

ENTRYPOINT ["/strimzi-canary"]
