FROM rockylinux:9
COPY bin/0.0.3/linux/amd64/gotfhub /usr/local/bin
RUN yum update -y && \
    chmod +x /usr/local/bin/gotfhub
ENTRYPOINT [ "/usr/local/bin/gotfhub" ]