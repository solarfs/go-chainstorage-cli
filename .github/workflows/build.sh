#!/bin/sh
set -ex


token="github_api_token"

for goos in darwin windows linux
do
	dirname="gcscmd_${GITHUB_REF_NAME}_${goos}_amd64"
	filename="$dirname"

	rm -rf $dirname $dirname.zip

	if [ "$goos" == "windows" ];then
		filename=${filename}.exe
	fi

	mkdir -p $dirname

	GOOS=$goos GOARCH=amd64 CGO_ENABLED=0 go build -o $dirname/$filename #-ldflags "-X main.version=tag"
	cp -R LICENSE README.md config.toml.sample MANUAL.md CHANGELOG.md changelogs/ $dirname/

	zip -r $dirname.zip $dirname
        pwd
        /bin/bash .github/workflows/upload-github-release-asset.sh github_api_token=${FM_CICD_TOKEN} owner=solarfs repo=go-chainstorage-cli tag=${GITHUB_REF_NAME} filename=./$dirname.zip
	rm -rf $dirname.zip $dirname
done



