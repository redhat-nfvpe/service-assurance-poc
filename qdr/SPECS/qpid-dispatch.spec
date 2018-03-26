# Define pkgdocdir for releases that don't define it already
%{!?_pkgdocdir: %global _pkgdocdir %{_docdir}/%{name}-%{version}}

# determine whether to install systemd or sysvinit scripts
%if 0%{?fedora} || 0%{?rhel} >= 7

%global _use_sysvinit 0
%global _use_systemd 1

%else

%global _use_sysvinit 1
%global _use_systemd 0

%endif

%global proton_minimum_version 0.18.0
%global libwebsockets_minimum_version 2.1.0

Name:          qpid-dispatch
Version:       1.0.1
Release:       1%{?dist}
Summary:       Dispatch router for Qpid
License:       ASL 2.0
URL:           http://qpid.apache.org/
Source0:       http://www.apache.org/dist/qpid/dispatch/%{version}/qpid-dispatch-%{version}.tar.gz

Source2:       licenses.xml

%global _pkglicensedir %{_licensedir}/%{name}-%{version}
%{!?_licensedir:%global license %doc}
%{!?_licensedir:%global _pkglicensedir %{_pkgdocdir}}

%if 0%{?rhel} && 0%{?rhel} >= 7
ExcludeArch: i686
%endif

%if 0%{?rhel} && 0%{?rhel} <= 6
Source1:       docs.tar.gz
%endif

Patch1:        0001-NO-JIRA-Systemd-control-file-for-qdrouterd.patch
Patch2:        0001-Added-etc-qdrouterd.patch

BuildRequires: qpid-proton-c-devel >= %{proton_minimum_version}
BuildRequires: python-devel
BuildRequires: cmake
BuildRequires: python-qpid-proton >= %{proton_minimum_version}
BuildRequires: openssl-devel
%if 0%{?fedora} && 0%{?fedora} >= 25
BuildRequires: libwebsockets-devel >= %{libwebsockets_minimum_version}
%endif
# Missing dependency on RHEL 6: asciidoc >= 8.6.8
# asciidoc-8.4.5-4.1.el6 does not work for man pages
%if 0%{?fedora} || (0%{?rhel} && 0%{?rhel} >= 7)
BuildRequires: asciidoc >= 8.6.8
BuildRequires: systemd
%endif


%description
A lightweight message router, written in C and built on Qpid Proton, that provides flexible and scalable interconnect between AMQP endpoints or between endpoints and brokers.


%package router
Summary:  The Qpid Dispatch Router executable
Obsoletes: libqpid-dispatch
Obsoletes: libqpid-dispatch-devel
Requires:  qpid-proton-c%{?_isa} >= %{proton_minimum_version}
Requires:  python
Requires:  python-qpid-proton >= %{proton_minimum_version}
%if %{_use_systemd}
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd
%endif
%if 0%{?fedora} && 0%{?fedora} >= 25
Requires: libwebsockets >= %{libwebsockets_minimum_version}
%endif


%description router
%{summary}.


%files router
%license %{_pkglicensedir}/LICENSE
%license %{_pkglicensedir}/licenses.xml
%{_sbindir}/qdrouterd
%config(noreplace) %{_sysconfdir}/qpid-dispatch/qdrouterd.conf
%config(noreplace) %{_sysconfdir}/sasl2/qdrouterd.conf
%{_exec_prefix}/lib/qpid-dispatch
%exclude %{_exec_prefix}/lib/qpid-dispatch/tests
%{python_sitelib}/qpid_dispatch_site.py*
%{python_sitelib}/qpid_dispatch
   
%if "%{python_version}" >= "2.6"
%{python_sitelib}/qpid_dispatch-*.egg-info
%endif

%if %{_use_systemd}

%{_unitdir}/qdrouterd.service

%else

%{_initrddir}/qdrouterd
%attr(755,qdrouterd,qdrouterd) %dir %{_localstatedir}/run/qpid-dispatch

%endif

%{_mandir}/man5/qdrouterd.conf.5*
%{_mandir}/man8/qdrouterd.8*

%pre router
getent group qdrouterd >/dev/null || groupadd -r qdrouterd
getent passwd qdrouterd >/dev/null || \
  useradd -r -M -g qdrouterd -d %{_localstatedir}/lib/qdrouterd -s /sbin/nologin \
    -c "Owner of Qdrouterd Daemons" qdrouterd
exit 0


%if %{_use_systemd}

%post router
/sbin/ldconfig
%systemd_post qdrouterd.service

%preun router
%systemd_preun qdrouterd.service

%postun router
/sbin/ldconfig
%systemd_postun_with_restart qdrouterd.service

%endif

%if %{_use_sysvinit}

%post router
/sbin/ldconfig
/sbin/chkconfig --add qdrouterd

%preun router
if [ $1 -eq 0 ]; then
    /sbin/service qdrouterd stop >/dev/null 2>&1
    /sbin/chkconfig --del qdrouterd
fi

%postun router
/sbin/ldconfig
if [ "$1" -ge "1" ]; then
    /sbin/service qdrouterd condrestart >/dev/null 2>&1
fi

%endif


%package docs
Summary:   Documentation for the Qpid Dispatch router
BuildArch: noarch
Obsoletes:  qpid-dispatch-router-docs

%description docs
%{summary}.


%files docs
%doc %{_pkgdocdir}
%license %{_pkglicensedir}/LICENSE
%license %{_pkglicensedir}/licenses.xml


%package tools
Summary: Tools for the Qpid Dispatch router
Requires: python-qpid-proton >= %{proton_minimum_version}
Requires: qpid-dispatch-router%{?_isa} = %{version}-%{release}


%description tools
%{summary}.


%files tools
%{_bindir}/qdstat
%{_bindir}/qdmanage

%{_mandir}/man8/qdstat.8*
%{_mandir}/man8/qdmanage.8*


%prep
%setup -q

%if %{_use_systemd}

%patch1 -p1

%else

%patch2 -p1

%endif

%if 0%{?rhel} && 0%{?rhel} <= 6
mkdir pre_built
cd pre_built
tar xvzpf %{SOURCE1}  -C .
%endif

%build
export DOCS=ON
%if 0%{?rhel} && 0%{?rhel} <= 6
export DOCS=OFF
%endif
%if 0%{?fedora} && 0%{?fedora} >= 25
export LIBWEBSOCKETS=ON
%else
export LIBWEBSOCKETS=OFF
%endif
%cmake -DDOC_INSTALL_DIR=%{?_pkgdocdir} \
       -DCMAKE_BUILD_TYPE=RelWithDebInfo \
       -DUSE_SETUP_PY=0 \
       -DQD_DOC_INSTALL_DIR=%{_pkgdocdir} \
       "-DBUILD_DOCS=$DOCS" \
       -DCMAKE_SKIP_RPATH:BOOL=OFF \
      "-DUSE_LIBWEBSOCKETS=$LIBWEBSOCKETS" \
       .
make doc


%install
%make_install

%if %{_use_systemd}

install -dm 755 %{buildroot}%{_unitdir}
install -pm 644 %{_builddir}/qpid-dispatch-%{version}/etc/qdrouterd.service \
                %{buildroot}%{_unitdir}

%else

install -dm 755 %{buildroot}%{_initrddir}
install -pm 755 %{_builddir}/qpid-dispatch-%{version}/etc/qdrouterd \
                %{buildroot}%{_initrddir}

%endif

%if 0%{?rhel} && 0%{?rhel} <= 6
install -dm 755 %{buildroot}%{_mandir}/man5
install -dm 755 %{buildroot}%{_mandir}/man8
install -pm 644 %{_builddir}/qpid-dispatch-%{version}/pre_built/man/man5/* %{buildroot}%{_mandir}/man5/
install -pm 644 %{_builddir}/qpid-dispatch-%{version}/pre_built/man/man8/* %{buildroot}%{_mandir}/man8/
cp -a           %{_builddir}/qpid-dispatch-%{version}/pre_built/doc/qpid-dispatch-%{version}/*  \
                %{buildroot}%{_pkgdocdir}/
%endif

install -dm 755 %{buildroot}/var/run/qpid-dispatch

%if 0%{?rhel} && 0%{?rhel} <= 6
install -pm 644 %{SOURCE2} %{buildroot}%{_pkgdocdir}
%else
install -dm 755 %{buildroot}%{_pkglicensedir}
install -pm 644 %{SOURCE2} %{buildroot}%{_pkglicensedir}
install -pm 644 %{buildroot}%{_pkgdocdir}/LICENSE %{buildroot}%{_pkglicensedir}
rm -f %{buildroot}%{_pkgdocdir}/LICENSE
%endif

rm -f  %{buildroot}/%{_includedir}/qpid/dispatch.h
rm -fr %{buildroot}/%{_includedir}/qpid/dispatch
rm -fr %{buildroot}/%{_datarootdir}/qpid-dispatch/console/stand-alone

%post -p /sbin/ldconfig

%postun -p /sbin/ldconfig


%changelog
* Tue Nov 21 2017 Irina Boverman <iboverma@redhat.com> - 1.0.0-1
- Rebased to 1.0.0
- Added DISPATCH-881 fix

* Thu Nov 16 2017 Irina Boverman <iboverma@redhat.com> - 0.8.0-6
- Rebuilt against qpid-proton 0.18.1

* Mon Aug 14 2017 Irina Boverman <iboverma@redhat.com> - 0.8.0-5
- Added fix for DISPATCH-727

* Mon Aug 14 2017 Fedora Release Engineering <releng@fedoraproject.org> - 0.8.0-4
- Rebuilt against latest version of libwebsockets

* Thu Aug 03 2017 Fedora Release Engineering <releng@fedoraproject.org> - 0.8.0-3
- Rebuilt for https://fedoraproject.org/wiki/Fedora_27_Binutils_Mass_Rebuild

* Thu Jul 27 2017 Fedora Release Engineering <releng@fedoraproject.org> - 0.8.0-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_27_Mass_Rebuild

* Thu May 25 2017 Irina Boverman <iboverma@redhat.com> - 0.8.0-1
- Rebased to 0.8.0

* Wed Feb 22 2017  Irina Boverman <iboverma@redhat.com> - 0.7.0-1
- Rebased to 0.7.0
- Rebuilt against qpid-proton 0.17.0

* Sat Feb 11 2017 Fedora Release Engineering <releng@fedoraproject.org> - 0.6.1-5
- Rebuilt for https://fedoraproject.org/wiki/Fedora_26_Mass_Rebuild

* Wed Feb  1 2017 Irina Boverman <iboverma@redhat.com> - 0.6.1-4
- Updated "Requires: python-qpid-proton" to use >= %{proton_minimum_version}

* Thu Sep  8 2016 Irina Boverman <iboverma@redhat.com> - 0.6.1-3
- Rebuilt against qpid-proton 0.14.0

* Tue Aug 23 2016 Irina Boverman <iboverma@redhat.com> - 0.6.1-2
- Obsoleted libqpid-dispatch-devel

* Wed Aug 17 2016 Irina Boverman <iboverma@redhat.com> - 0.6.1-1
- Rebased to 0.6.1
- Corrected doc package build process

* Tue Jul 19 2016 Fedora Release Engineering <rel-eng@lists.fedoraproject.org> - 0.6.0-2
- https://fedoraproject.org/wiki/Changes/Automatic_Provides_for_Python_RPM_Packages

* Fri Jun 24 2016 Irina Boverman <iboverma@redhat.com> - 0.6.0-1
- Rebased to 0.6.0
- Rebuilt against qpid-proton 0.13.0-1
- Changed qpid-dispatch-router-docs to qpid-dispatch-docs

* Wed Mar 23 2016 Irina Boverman <iboverma@redhat.com> - 0.5-3
- Rebuilt against proton 0.12.1

* Thu Feb 04 2016 Fedora Release Engineering <releng@fedoraproject.org> - 0.5-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_24_Mass_Rebuild

* Wed Sep 16 2015 Irina Boverman <iboverma@redhat.com> - 0.5-1
- Rebased to qpid dispatch 0.5

* Thu Jun 18 2015 Fedora Release Engineering <rel-eng@lists.fedoraproject.org> - 0.4-3
- Rebuilt for https://fedoraproject.org/wiki/Fedora_23_Mass_Rebuild

* Wed May 27 2015 Darryl L. Pierce <dpierce@redhat.com> - 0.4-2
- Create the local state directory explicity on SysVInit systems.

* Tue Apr 21 2015 Darryl L. Pierce <dpierce@rehdat.com> - 0.4-1
- Rebased on Dispatch 0.4.
- Changed username for qdrouterd to be qdrouterd.

* Tue Feb 24 2015 Darryl L. Pierce <dpierce@redhat.com> - 0.3-4
- Changed SysVInit script to properly name qdrouterd as the service to start.

* Fri Feb 20 2015 Darryl L. Pierce <dpierce@redhat.com> - 0.3-3
- Update inter-package dependencies to include release as well as version.

* Wed Feb 11 2015 Darryl L. Pierce <dpierce@redhat.com> - 0.3-2
- Disabled building documentation due to missing pandoc-pdf on EL6.
- Disabled daemon setgid.
- Fixes to accomodate Python 2.6 on EL6.
- Removed implicit dependency on python-qpid-proton in qpid-dispatch-router.

* Tue Jan 27 2015 Darryl L. Pierce <dpierce@redhat.com> - 0.3-1
- Rebased on Dispatch 0.3.
- Increased the minimum Proton version needed to 0.8.
- Moved all tests to the -devel package.
- Ensure executable bit turned off on systemd file.
- Set the location of installed documentation.

* Thu Nov 20 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.2-9
- Fixed a merge issue that resulted in two patches not being applied.
- Resolves: BZ#1165691

* Wed Nov 19 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.2-8
- DISPATCH-75 - Removed reference to qdstat.conf from qdstat manpage.
- Include systemd service file for EPEL7 packages.
- Brought systemd support up to current Fedora packaging guidelines.
- Resolves: BZ#1165691
- Resolves: BZ#1165681

* Sun Aug 17 2014 Fedora Release Engineering <rel-eng@lists.fedoraproject.org> - 0.2-7
- Rebuilt for https://fedoraproject.org/wiki/Fedora_21_22_Mass_Rebuild

* Wed Jul  9 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.2-6
- Removed intro-package comments which can cause POSTUN warnings.
- Added dependency on libqpid-dispatch from qpid-dispatch-tools.

* Wed Jul  2 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.2-5
- Fixed the path for the configuration file.
- Resolves: BZ#1115416

* Sun Jun 08 2014 Fedora Release Engineering <rel-eng@lists.fedoraproject.org> - 0.2-4
- Rebuilt for https://fedoraproject.org/wiki/Fedora_21_Mass_Rebuild

* Fri May 30 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.2-3
- Fixed build type to be RelWithDebInfo

* Tue Apr 22 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.2-2
- Fixed merging problems across Fedora and EPEL releases.

* Tue Apr 22 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.2-1
- Rebased on Qpid Dispatch 0.2.

* Wed Feb  5 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.1-4
- Fixed path to configuration in qpid-dispatch.service file.
- Added requires from qpid-dispatch-tools to python-qpid-proton.

* Thu Jan 30 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.1-3
- Fix build system to not discard CFLAGS provided by Fedora
- Resolves: BZ#1058448
- Simplified the specfile to be used across release targets.

* Fri Jan 24 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.1-2
- First release for Fedora.
- Resolves: BZ#1055721

* Thu Jan 23 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.1-1.2
- Put all subpackage sections above prep/build/install.
- Removed check and clean sections.
- Added remaining systemd macros.
- Made qpid-dispatch-router-docs a noarch package.

* Wed Jan 22 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.1-1.1
- Added the systemd macros for post/preun/postun
- Moved prep/build/install/check/clean above package definitions.

* Mon Jan 20 2014 Darryl L. Pierce <dpierce@redhat.com> - 0.1-1
- Initial packaging of the codebase.
