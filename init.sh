#!/bin/bash

cd examples && make && cd -
BASE_IMAGE=netutils make container
cp examples/demo.db examples/pb_data/data.db
cd examples && docker compose up