#!/bin/bash

make
./.output/presilo -o ./.temp/ -l go -m main ./samples/wardrobe.json
./.output/presilo -o ./.temp/ -l js -m main ./samples/wardrobe.json
./.output/presilo -o ./.temp/ -l java -m foo ./samples/wardrobe.json

./.output/presilo -o ./.temp/ -l go -m main ./samples/person.json
./.output/presilo -o ./.temp/ -l js -m main ./samples/person.json
./.output/presilo -o ./.temp/ -l java -m foo ./samples/person.json

./.output/presilo -o ./.temp/ -l go -m main ./samples/car.json
./.output/presilo -o ./.temp/ -l js -m main ./samples/car.json
./.output/presilo -o ./.temp/ -l java -m foo ./samples/car.json

pushd ./.temp/
go build .

mkdir -p foo
mv *.java ./foo/
cd ./foo
javac $(find . -name "*.java")
popd
