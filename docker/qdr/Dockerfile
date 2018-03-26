FROM fedora:27
COPY ./qdr/RPMS/*x86_64.rpm /root/
RUN dnf install /root/qpid*.rpm -y && dnf clean all && rm -f /root/*.rpm
ENTRYPOINT ["/usr/sbin/qdrouterd"]
