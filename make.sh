#!/usr/bin/env sh
VER=0.2

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
