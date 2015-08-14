#!/bin/bash

make
./.output/presilo -o ./.temp/ -l go -m main ./samples/wardrobe.json
./.output/presilo -o ./.temp/ -l js -m main ./samples/wardrobe.json

./.output/presilo -o ./.temp/ -l go -m main ./samples/person.json
./.output/presilo -o ./.temp/ -l js -m main ./samples/person.json

./.output/presilo -o ./.temp/ -l go -m main ./samples/car.json
./.output/presilo -o ./.temp/ -l js -m main ./samples/car.json
