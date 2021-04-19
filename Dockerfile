FROM scratch

ADD cmd/target/strimzi-canary ./

ENTRYPOINT ["/strimzi-canary"]