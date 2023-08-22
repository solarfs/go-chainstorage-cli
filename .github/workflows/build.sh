#!/bin/sh
set -ex


token="github_api_token"

#tag=$(git describe --tags `git rev-list --tags --max-count=1`)


#for goos in darwin windows linux
for goos in linux
do
	dirname="gcscmd_tag_${goos}_amd64"
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
#	sh  upload-github-release-asset.sh github_api_token=$token owner=solarfs repo=go-chainstorage-cli tag=$tag filename=./$dirname.zip
        /bin/bash .github/workflows/upload-github-release-asset.sh github_api_token=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 owner=solarfs repo=go-chainstorage-cli tag=v0.0.9test filename=./$dirname.zip
	rm -rf $dirname.zip $dirname
done
git --version

echo ${FM_USERNAME}


