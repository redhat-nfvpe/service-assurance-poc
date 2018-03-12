FROM fedora:27
LABEL maintainer="admin@nfvpe.site"
RUN dnf install qpid-proton-c -y && yum clean all
ADD main /
EXPOSE 8081
ENTRYPOINT ["/main"]
CMD ["-mhost=0.0.0.0", "-mport=8081", "-usesample=true", "-h=10", "-p=100", "-t=-1"]
