Summary:        PostgreSQL database engine
Name:           postgresql96
Version:        9.6.21
Release:        1%{?dist}
License:        PostgreSQL
URL:            www.postgresql.org
Group:          Applications/Databases
Vendor:         VMware, Inc.
Distribution:   Photon

Source0:        http://ftp.postgresql.org/pub/source/v%{version}/%{name}-%{version}.tar.bz2
%define sha1    postgresql=e24333824d361968958613f546ae06011d9d1dfc
# Common libraries needed
BuildRequires:  krb5-devel
BuildRequires:  libxml2-devel
BuildRequires:  openldap
BuildRequires:  perl
BuildRequires:  readline-devel
BuildRequires:  openssl-devel
BuildRequires:  zlib-devel
BuildRequires:  tzdata
BuildRequires:  bzip2
BuildRequires:  sudo
Requires:       krb5
Requires:       libxml2
Requires:       openldap
Requires:       openssl
Requires:       readline
Requires:       zlib
Requires:       tzdata
Requires:       bzip2
Requires:       sudo

Requires:   %{name}-libs = %{version}-%{release}

%description
PostgreSQL is an object-relational database management system.

%package libs
Summary:    Libraries for use with PostgreSQL
Group:      Applications/Databases

%description libs
The postgresql-libs package provides the essential shared libraries for any
PostgreSQL client program or interface. You will need to install this package
to use any other PostgreSQL package or any clients that need to connect to a
PostgreSQL server.

%package        devel
Summary:        Development files for postgresql.
Group:          Development/Libraries
Requires:       postgresql = %{version}-%{release}

%description    devel
The postgresql-devel package contains libraries and header files for
developing applications that use postgresql.

%prep
%setup -q

%build
ls -la
sed -i '/DEFAULT_PGSOCKET_DIR/s@/tmp@/run/postgresql@' src/include/pg_config_manual.h &&
./configure \
    --prefix=/usr/local/pg96 \
    --with-includes=/usr/local/pg96/include \
    --with-libraries=/usr/local/pg96/lib \
    --datarootdir=/usr/local/pg96/share \
    --enable-thread-safety \
    --with-ldap \
    --with-libxml \
    --with-openssl \
    --with-gssapi \
    --with-readline \
    --with-system-tzdata=%{_datadir}/zoneinfo \
    --docdir=/usr/local/pg96/doc/postgresql
make %{?_smp_mflags}
cd contrib && make %{?_smp_mflags}

%install
[ %{buildroot} != "/"] && rm -rf %{buildroot}/*
make install DESTDIR=%{buildroot}
cd contrib && make install DESTDIR=%{buildroot}

%{_fixperms} %{buildroot}/*

%check
sed -i '2219s/",/  ; EXIT_STATUS=$? ; sleep 5 ; exit $EXIT_STATUS",/g'  src/test/regress/pg_regress.c
chown -Rv nobody .
sudo -u nobody -s /bin/bash -c "PATH=$PATH make -k check"

%post   -p /sbin/ldconfig
%postun -p /sbin/ldconfig
%clean
rm -rf %{buildroot}/*

%files
%defattr(-,root,root)
/usr/local/pg96/bin/initdb
/usr/local/pg96/bin/oid2name
/usr/local/pg96/bin/pg_archivecleanup
/usr/local/pg96/bin/pg_basebackup
/usr/local/pg96/bin/pg_controldata
/usr/local/pg96/bin/pg_ctl
/usr/local/pg96/bin/pg_receivexlog
/usr/local/pg96/bin/pg_recvlogical
/usr/local/pg96/bin/pg_resetxlog
/usr/local/pg96/bin/pg_rewind
/usr/local/pg96/bin/pg_standby
/usr/local/pg96/bin/pg_test_fsync
/usr/local/pg96/bin/pg_test_timing
/usr/local/pg96/bin/pg_upgrade
/usr/local/pg96/bin/pg_xlogdump
/usr/local/pg96/bin/pgbench
/usr/local/pg96/bin/postgres
/usr/local/pg96/bin/postmaster
/usr/local/pg96/bin/vacuumlo
/usr/local/pg96/share/postgresql/*
/usr/local/pg96/lib/postgresql/*
/usr/local/pg96/doc/postgresql/extension/*.example
%exclude /usr/local/pg96/share/postgresql/pg_service.conf.sample
%exclude /usr/local/pg96/share/postgresql/psqlrc.sample

%files libs
/usr/local/pg96/bin/clusterdb
/usr/local/pg96/bin/createdb
/usr/local/pg96/bin/createlang
/usr/local/pg96/bin/createuser
/usr/local/pg96/bin/dropdb
/usr/local/pg96/bin/droplang
/usr/local/pg96/bin/dropuser
/usr/local/pg96/bin/ecpg
/usr/local/pg96/bin/pg_config
/usr/local/pg96/bin/pg_dump
/usr/local/pg96/bin/pg_dumpall
/usr/local/pg96/bin/pg_isready
/usr/local/pg96/bin/pg_restore
/usr/local/pg96/bin/psql
/usr/local/pg96/bin/reindexdb
/usr/local/pg96/bin/vacuumdb
/usr/local/pg96/lib/libecpg*.so.*
/usr/local/pg96/lib/libpgtypes*.so.*
/usr/local/pg96/lib/libpq*.so.*
/usr/local/pg96/share/postgresql/pg_service.conf.sample
/usr/local/pg96/share/postgresql/psqlrc.sample

%files devel
%defattr(-,root,root)
/usr/local/pg96/include/*
/usr/local/pg96/lib/pkgconfig/*
/usr/local/pg96/lib/libecpg*.so
/usr/local/pg96/lib/libpgtypes*.so
/usr/local/pg96/lib/libpq*.so
/usr/local/pg96/lib/libpgcommon.a
/usr/local/pg96/lib/libpgfeutils.a
/usr/local/pg96/lib/libpgport.a
/usr/local/pg96/lib/libpq.a
/usr/local/pg96/lib/libecpg.a
/usr/local/pg96/lib/libecpg_compat.a
/usr/local/pg96/lib/libpgtypes.a

%changelog
*   Thu Feb 18 2021 Michael Paquier <mpaquier@vmware.com> 9.6.21-1
-   Updated to version 9.6.21
*   Fri Nov 20 2020 Dweep Advani <dadvani@vmware.com> 9.6.20-1
-   Upgrading to 9.6.20 for addressing multiple CVEs
*   Tue Sep 01 2020 Dweep Advani <dadvani@vmware.com> 9.6.19-1
-   Upgrading to 9.6.19 for addressing multiple CVEs
*   Fri Apr 03 2020 Anisha Kumari <kanisha@vmware.com> 9.6.14-3
-   Added patch to fix CVE-2020-1720
*   Mon Nov 18 2019 Prashant S Chauhan <psinghchauha@vmware.com> 9.6.14-2
-   Added patch to fix CVE-2019-10208
*   Tue Jun 25 2019 Siju Maliakkal <smaliakkal@vmware.com> 9.6.14-1
-   Upgrade to 9.6.14 for CVE-2019-10164
*   Tue Aug 21 2018 Keerthana K <keerthanak@vmware.com> 9.6.10-1
-   Updated to version 9.6.10 to fix CVE-2018-10915, CVE-2018-10925.
*   Tue May 29 2018 Xiaolin Li <xiaolinl@vmware.com> 9.6.9-1
-   Updated to version 9.6.9
*   Tue Mar 27 2018 Dheeraj Shetty <dheerajs@vmware.com> 9.6.8-1
-   Updated to version 9.6.8 to fix CVE-2018-1058
*   Mon Feb 12 2018 Dheeraj Shetty <dheerajs@vmware.com> 9.6.7-1
-   Updated to version 9.6.7
*   Mon Nov 27 2017 Xiaolin Li <xiaolinl@vmware.com> 9.6.6-1
-   Updated to version 9.6.6
*   Fri Sep 08 2017 Xiaolin Li <xiaolinl@vmware.com> 9.6.5-1
-   Updated to version 9.6.5
*   Tue Aug 15 2017 Xiaolin Li <xiaolinl@vmware.com> 9.6.4-1
-   Updated to version 9.6.4
*   Thu Aug 10 2017 Rongrong Qiu <rqiu@vmware.com> 9.6.3-3
-   add sleep 5 when initdb in make check for bug 1900371
*   Wed Jul 05 2017 Divya Thaluru <dthaluru@vmware.com> 9.6.3-2
-   Added postgresql-devel
*   Tue Jun 06 2017 Divya Thaluru <dthaluru@vmware.com> 9.6.3-1
-   Upgraded to 9.6.3
*   Mon Apr 03 2017 Rongrong Qiu <rqiu@vmware.com> 9.6.2-1
-   Upgrade to 9.6.2 for Photon upgrade bump
*   Thu Dec 15 2016 Xiaolin Li <xiaolinl@vmware.com> 9.5.3-6
-   Applied CVE-2016-5423.patch
*   Thu Nov 24 2016 Alexey Makhalov <amakhalov@vmware.com> 9.5.3-5
-   Required krb5-devel.
*   Mon Oct 03 2016 ChangLee <changLee@vmware.com> 9.5.3-4
-   Modified %check
*   Thu May 26 2016 Xiaolin Li <xiaolinl@vmware.com> 9.5.3-3
-   Add tzdata to buildrequires and requires.
*   Tue May 24 2016 Priyesh Padmavilasom <ppadmavilasom@vmware.com> 9.5.3-2
-   GA - Bump release of all rpms
*   Fri May 20 2016 Divya Thaluru <dthaluru@vmware.com> 9.5.3-1
-   Updated to version 9.5.3
*   Wed Apr 13 2016 Michael Paquier <mpaquier@vmware.com> 9.5.2-1
-   Updated to version 9.5.2
*   Tue Feb 23 2016 Xiaolin Li <xiaolinl@vmware.com> 9.5.1-1
-   Updated to version 9.5.1
*   Thu Jan 21 2016 Xiaolin Li <xiaolinl@vmware.com> 9.5.0-1
-   Updated to version 9.5.0
*   Thu Aug 13 2015 Divya Thaluru <dthaluru@vmware.com> 9.4.4-1
-   Update to version 9.4.4.
*   Mon Jul 13 2015 Alexey Makhalov <amakhalov@vmware.com> 9.4.1-2
-   Exclude /usr/lib/debug
*   Fri May 15 2015 Sharath George <sharathg@vmware.com> 9.4.1-1
-   Initial build. First version
