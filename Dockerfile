FROM rockylinux:9
RUN yum update -y
COPY dist/bin/0.0.3/linux/amd64/gotfhub /usr/local/bin