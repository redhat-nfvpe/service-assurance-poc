#!/bin/bash -ex
#
if [ "$VERSION" = "latest" ] || [ "$VERSION" = "devel" ]; then
  echo "VERSION is $VERSION..."
else
  echo "VERSION $VERSION is not supported"
fi
#
home=$PWD
dnf update -y
# for proton <= 0.17.0 (latest now is 0.18.0), use compat-openssl10, otherwise use openssl
openssl=openssl
#
# Install builod dependencies
#
dnf install -y git gcc gcc-c++ cmake make swig pkgconfig libuuid-devel python-devel python2-devel python3-devel cyrus-sasl-devel libwebsockets-devel ${openssl}-devel
#
# Build and install proton bits
#
git clone git://git.apache.org/qpid-proton.git
cd qpid-proton
if [ "$VERSION" = "latest" ]; then
  git checkout $QPID_PROTON_VERSION
fi
cmake -DBUILD_CPP=OFF -DSYSINSTALL_BINDINGS=ON -DSYSINSTALL_PYTHON=ON -DCMAKE_INSTALL_PREFIX=/usr .
make
make install
cd $home
#
# Build and install dispatch bits
#
git clone git://git.apache.org/qpid-dispatch.git
cd qpid-dispatch
if [ "$VERSION" = "latest" ]; then
  git checkout $QPID_DISPATCH_VERSION
fi
cmake -DUSE_LIBWEBSOCKETS=ON "-DCMAKE_C_FLAGS= -Wno-error=implicit-function-declaration" -DCMAKE_INSTALL_PREFIX=/usr .
make
make install
#
# Remove build dependencies
#
#dnf remove -y git gcc gcc-c++ cmake make swig pkgconfig libuuid-devel python-devel python2-devel python3-devel cyrus-sasl-devel libwebsockets-devel ${openssl}-devel
dnf remove -y git swig pkgconfig libuuid-devel cyrus-sasl-devel libwebsockets-devel ${openssl}-devel
#
# Install run time dependencies
#
dnf install -y python libwebsockets cyrus-sasl-plain cyrus-sasl-gssapi cyrus-sasl-md5 cyrus-sasl-lib ${openssl}
cd $home
#
# Cleanup
#
rm -fr qpid-proton qpid-dispatch
dnf clean all

# Install golang
rpm --import https://mirror.go-repo.io/fedora/RPM-GPG-KEY-GO-REPO
curl -s https://mirror.go-repo.io/fedora/go-repo.repo | tee /etc/yum.repos.d/go-repo.repo
dnf install -y golang git

export GOPATH=/opt/go
if [ ! -d $GOPATH ]; then
	mkdir -p $GOPATH
fi

go get github.com/redhat-nfvpe/service-assurance-poc
cd $GOPATH/src/github.com/redhat-nfvpe/service-assurance-poc
go get .
go get ./metrics
go get ./events
go get ./tools/sa-bench/.
go build .
go build -o metric-listener metrics/main.go
go build -o event-listener events/main.go
go build -o sa-bench tools/sa-bench/main.go
exit
