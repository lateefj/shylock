=================
Shylock Packaging
=================


The packaging systems goal is to package as many versions systems as possible. 

Currently there is a render.py script that generates rpm and deb configuration files. Before building any linux binaries distributions first have to build a linux binary using the make default.

```bash
make
```

Then each packaging system needs to build its own file which should end up in the build directory.

RPM
---

Vagrant file which should build an rpm into the build directory as build/shylock.rpm.

```bash
vagrant up
```

DEB
---

Vagrant file which should build an deb into the build directory as build/shylock.deb.

```bash
vagrant up
```

