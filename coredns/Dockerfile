FROM ubuntu
COPY ./coredns/coredns /coredns
RUN chmod +x /coredns
EXPOSE 53 53/udp
ENTRYPOINT ["/coredns"]