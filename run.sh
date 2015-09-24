#!/bin/bash

make
./.output/presilo -o ./.temp/go/ -l go -m foo ./samples/wardrobe.json
./.output/presilo -o ./.temp/js/ -l js -m foo ./samples/wardrobe.json
./.output/presilo -o ./.temp/java/ -l java -m foo ./samples/wardrobe.json
./.output/presilo -o ./.temp/cs/ -l cs -m foo ./samples/wardrobe.json
./.output/presilo -o ./.temp/rb/ -l rb -m foo ./samples/wardrobe.json
./.output/presilo -o ./.temp/py/ -l py -m foo ./samples/wardrobe.json
./.output/presilo -o ./.temp/mysql/ -l mysql -m foo ./samples/wardrobe.json

./.output/presilo -o ./.temp/go/ -l go -m foo ./samples/person.json
./.output/presilo -o ./.temp/js/ -l js -m foo ./samples/person.json
./.output/presilo -o ./.temp/java/ -l java -m foo ./samples/person.json
./.output/presilo -o ./.temp/cs/ -l cs -m foo ./samples/person.json
./.output/presilo -o ./.temp/rb/ -l rb -m foo ./samples/person.json
./.output/presilo -o ./.temp/py/ -l py -m foo ./samples/person.json
./.output/presilo -o ./.temp/mysql/ -l mysql -m foo ./samples/person.json

./.output/presilo -o ./.temp/go/ -l go -m foo ./samples/car.json
./.output/presilo -o ./.temp/js/ -l js -m foo ./samples/car.json
./.output/presilo -o ./.temp/java/ -l java -m foo ./samples/car.json
./.output/presilo -o ./.temp/cs/ -l cs -m foo ./samples/car.json
./.output/presilo -o ./.temp/rb/ -l rb -m foo ./samples/car.json
./.output/presilo -o ./.temp/py/ -l py -m foo ./samples/car.json
./.output/presilo -o ./.temp/mysql/ -l mysql -m foo ./samples/car.json

pushd ./.temp/go
go build .
popd
pushd ./.temp/java
javac $(find . -name "*.java")
popd
pushd ./.temp/cs/
mcs $(find . -name "*.cs") --parse
popd
pushd ./.temp/rb/
ruby $(find . -name "*.rb")
popd
pushd ./.temp/py/
python $(find . -name "*.py")
popd
