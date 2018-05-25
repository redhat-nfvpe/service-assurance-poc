# QPID Dispatch Router RPM Build

The `./buildit.sh` script will result in a set of RPMs that can be consumed by
`docker build` to create the `nfvpe/qpid-dispatch-router` container image.

## Requirements

It is expected that `mock` is installed on your system and that your current
user is in the `mock` system group.

# Building for CentOS

Building for CentOS will require additional effort since you'll need to
download the source RPMs from Fedora, and backport them to CentOS. This is
relatively straight forward when you use `mock` to build the dependencies, then
load them into the `chroot` with a custom root configuration file.

## Requirements

You'll need to install the following dependencies:

* `git`
* `mock`

Install the dependencies with the following command.

    sudo yum install git mock -y

## Creating a Build User

Create the `mockbuild` user and assign it to the `mock` and `mockbuild` groups.

    sudo usermod -a -G mock -g mockbuild mockbuild

Switch to the `mockbuild` user and make it so you can SSH into that user with
your SSH key. This will be useful later when you need to copy in the dependency
SRPMs.

    sudo su - mockbuild
    mkdir ~/.ssh
    chmod 0700 ~/.ssh
    cat > ~/.ssh/authorized_keys <<EOF
    ...<your_ssh_public_key>...
    EOF
    chmod 0400 ~/.ssh/authorized_keys

## Download Dependency Source RPMs

From my Fedora 27 desktop, I downloaded the following source RPMs.

* qpid-proton
* libwebsockets
* libev

You can download them with the following commands.

    dnf download --enablerepo=updates-source --enablerepo=fedora-source qpid-proton.src
    dnf download --enablerepo=updates-source --enablerepo=fedora-source libwebsockets.src
    dnf download --enablerepo=updates-source --enablerepo=fedora-source libev.src

If your build machine is not the same you're downloading from, then upload them
now.

    scp *.src.rpm mockbuild@<build_machine>:

## Install Source RPM Dependencies

Install the source dependencies into your `~/rpmbuild/` directory with `rpm`.

    rpm -Uvh libev-4.24-4.fc27.src.rpm libwebsockets-2.3.0-2.fc27.src.rpm qpid-proton-0.18.1-1.fc27.src.rpm

## Create Local RPM Repo for Dependencies

When building with `mock` you pass in the root build system to use, such as
CentOS 7 x86_64, Fedora 27 x86_64, etc.

To make it easier to rebuild our libraries we create a local RPM repository on
our build system and make that available as a repository to the chroot when
mock is building our RPMs for us.

First, copy the default `centos-7-x86_64` build root configuration file into
your `~/rpmbuild/` directory.

    cp /etc/mock/centos-7-x86_64 ~/rpmbuild/

And now append the following text to that configuration file to expose the RPM
repository we'll create in a second.

    [local_deps]
    name=Local Dependencies
    baseurl=file:///home/mockbuild/rpmbuild/RPMS/
    enabled=1
    gpgcheck=0
    includepkgs=*.x86_64 *.noarch

> **NOTE**
>
> Make sure you apply the previous set of configuration ABOVE the trailing
> `"""` at the end of the file.

We'll come back and create our repository in a moment after we've built our
first set of RPMs (which requires no other dependencies).

## Building the RPMs

We're going to iterate through the following workflow.

* build dependency SRPM
* build dependency RPM
* create/update our local RPM repository with `createrepo_c`
* move onto next dependency (Goto 1)

The order of dependency build will be the following.

* libev
* libwebsockets
* qpid-proton
* qpid-dispatch

### `libev`

So let's start with building out our first library, `libev`.

    cd ~/rpmbuild/
    mock --buildsrpm --root centos-7-x86_64 --configdir=. \
        --spec=SPECS/libev.spec --sources=SOURCES/ --resultdir=SRPMS/
    mock --rebuild --root centos-7-x86_64 --configdir=. --resultdir=RPMS/ libev-4.24-4.el7.src.rpm

Now that we have our first library built (ideally everything has worked for
you, at least it did for me :)), we need to create our local repository.

    cd ~/rpmbuild/RPMS/
    createrepo_c .

With this done, let's move to the rest of our dependencies.

### `libwebsockets`

    cd ~/rpmbuild/
    mock --buildsrpm --root centos-7-x86_64 --configdir=. \
        --spec=SPECS/libwebsockets.spec --sources=SOURCES/ --resultdir=SRPMS/
    mock --rebuild --root centos-7-x86_64 --configdir=. --resultdir=RPMS/ libwebsockets-2.3.0-2.el7.src.rpm
    cd ~/rpmbuild/RPMS/
    createrepo_c .

### `qpid-proton`

    cd ~/rpmbuild/
    mock --buildsrpm --root centos-7-x86_64 --configdir=. \
        --spec=SPECS/qpid-proton.spec --sources=SOURCES/ --resultdir=SRPMS/
    mock --rebuild --root centos-7-x86_64 --configdir=. --resultdir=RPMS/ qpid-proton-0.18.1-1.el7.src.rpm
    cd ~/rpmbuild/RPMS/
    createrepo_c .

### `qpid-dispatch-router`

    cd ~
    git clone https://github.com/redhat-nfvpe/service-assurance-poc
    cd service-assurance-poc/qdr/
    cp -r ./SPECS/* ~/rpmbuild/SPECS
    cp -r ./SOURCES/* ~/rpmbuild/SOURCES/
    cd ~/rpmbuild/
    curl -sSL http://www.apache.org/dist/qpid/dispatch/1.0.1/qpid-dispatch-1.0.1.tar.gz -o SOURCES/qpid-dispatch-1.0.1.tar.gz

    mock --buildsrpm --root centos-7-x86_64 --configdir=. \
        --spec=SPECS/qpid-dispatch.spec --sources=SOURCES/ --resultdir=SRPMS/
    mock --rebuild --root centos-7-x86_64 --configdir=. --resultdir=RPMS/qpid-dispatch-1.0.1-1.el7.src.rpm
