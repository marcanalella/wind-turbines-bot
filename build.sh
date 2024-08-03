#!/bin/bash

echo "Insert build version (example 1.0):"
read version

echo "Preparing build: $version";
docker build -t wind-turbine-bot:$version .
docker login -u mcanalella
docker tag wind-turbine-bot:$version mcanalella/wind-turbine-bot:$version
docker push mcanalella/wind-turbine-bot:$version

exit 0
