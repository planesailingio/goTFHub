FROM rockylinux:9.3
ARG VERSION
COPY bin/$VERSION/linux/amd64/gotfhub /usr/local/bin
RUN yum update -y && \
    chmod +x /usr/local/bin/gotfhub
ENTRYPOINT [ "/usr/local/bin/gotfhub" ]