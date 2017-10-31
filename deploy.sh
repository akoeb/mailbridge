#!/bin/bash
# travis would not execute those things in a after_success block
# so ended up creating a deploy script
echo "Running deploy script"
goversion=$(go version)
if [ -z "$goversion" ]
then
  echo "Could not find go version">&2
  exit 1
fi

if echo $goversion | egrep -q '\sgo1\.7\.[0-9]\s' ; then
    echo "go version correct, running deploy"
    echo ${TRAVIS_COMMIT} > COMMIT
    docker build -t $REPO:$TRAVIS_BUILD_NUMBER -f Dockerfile .
    docker login -u $DOCKER_USER -p $DOCKER_PASS
    docker tag $REPO:$TRAVIS_BUILD_NUMBER $REPO:latest
    docker push $REPO:$TRAVIS_BUILD_NUMBER
    docker push $REPO:latest
fi
