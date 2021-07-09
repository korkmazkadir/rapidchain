#!/bin/bash

# builds registry service
cd ../cmd/registery/
env GOOS=linux GOARCH=amd64 go build .

cd -

# builds node
cd ../cmd/node/
env GOOS=linux GOARCH=amd64 go build .

cd -

# removes the artifacts folder
mkdir ./artifacts

mv ../cmd/registery/registery ./artifacts
cp ../cmd/registery/config.json ./artifacts
mv ../cmd/node/node ./artifacts
cp deploy-nodes.sh ./artifacts
