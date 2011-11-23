#!/usr/bin/env sh
VER=0.1-2

# version
echo "package main; var version = \"$VER\"" >src/tmsu/version.go

# compile
pushd src/tmsu
gomake
popd

# create archive
mkdir -p tmsu-$VER/bin
cp src/tmsu/tmsu tmsu-$VER/bin
cp LICENSE README tmsu-$VER
tar czf tmsu-$VER.tgz tmsu-$VER
rm -R tmsu-$VER
