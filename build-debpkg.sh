#!/bin/bash
sed -i -E "s#!THE_VERSION!#$1#" ./.debpkg/DEBIAN/control
dpkg-deb --build ./.debpkg .
