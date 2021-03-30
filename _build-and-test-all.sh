#! /bin/bash

set -e

dockerall="./gradlew ${DATABASE?}${MODE?}Compose"
dockerinfrastructure="./gradlew ${DATABASE?}${MODE?}infrastructureCompose"

./gradlew testClasses

${dockerall}Down -P removeContainers=true
${dockerinfrastructure}Up

cd customer-service
go build .
cd ..

echo 'show databases;' | ./mysql-cli.sh -i

${dockerall}Build
${dockerall}Up

echo 'show dbs' |  ./mongodb-cli.sh -i

./gradlew :end-to-end-tests:cleanTest :end-to-end-tests:test

${dockerall}Down -P removeContainers=true