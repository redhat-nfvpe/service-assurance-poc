#!/usr/bin/env bash
echo "--- Getting QPID Dispatch Router Sources"
curl -sSL http://www.apache.org/dist/qpid/dispatch/1.0.1/qpid-dispatch-1.0.1.tar.gz -o SOURCES/qpid-dispatch-1.0.1.tar.gz

echo "--- Building SRPM"
mock --root fedora-27-x86_64 --buildsrpm --resultdir=SRPMS/ --sources=SOURCES/ --spec SPECS/qpid-dispatch.spec

echo "--- Build RPM"
mock --root=fedora-27-x86_64 --rebuild --resultdir=RPMS/ SRPMS/qpid-dispatch-1.0.1-1.fc27.src.rpm
