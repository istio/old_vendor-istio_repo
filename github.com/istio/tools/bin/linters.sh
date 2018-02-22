#!/bin/bash
set -ex

SCRIPTPATH=$( cd "$(dirname "$0")" ; pwd -P )

WORKSPACE=$SCRIPTPATH/..

cd ${WORKSPACE}

GOOS=
GOARCH=

echo 'Installing gometalinter ....'
go get -u gopkg.in/alecthomas/gometalinter.v2
gometalinter=gometalinter.v2
$gometalinter --install
echo 'Gometalinter installed successfully ....'

echo 'Running gometalinter ....'
$gometalinter --config=./lintconfig_base.json ./...
echo 'gometalinter OK'
