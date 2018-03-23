FROM fedora:27
LABEL maintainer="admin@nfvpe.site"
RUN dnf install qpid-proton-c -y && dnf clean all
ADD main /
EXPOSE 8081
ENTRYPOINT ["/main"]
