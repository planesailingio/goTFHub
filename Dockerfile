FROM rockylinux:9
ARG VERSION
COPY bin/$VERSION/linux/amd64/gotfhub /usr/local/bin
RUN yum update -y && yum install -y git && \
    chmod +x /usr/local/bin/gotfhub
ENTRYPOINT [ "/usr/local/bin/gotfhub" ]