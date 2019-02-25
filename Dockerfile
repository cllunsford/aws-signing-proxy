FROM scratch
MAINTAINER uSwitch Cloud <cloud@uswitch.com>

COPY bin/aws-signing-proxy-linux-amd64 /aws-signing-proxy

EXPOSE 8080

ENTRYPOINT ["/aws-signing-proxy"]
