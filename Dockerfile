FROM rockylinux:9
RUN yum update -y
COPY . /app