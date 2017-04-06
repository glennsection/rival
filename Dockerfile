FROM alpine:latest

MAINTAINER Glenn Powell <glenn.powell@sectionstudios.com>

WORKDIR "/opt"

ADD .docker_build/bloodtales /opt/bin/bloodtales
ADD ./templates /opt/templates
ADD ./static /opt/static

CMD ["/opt/bin/bloodtales"]

