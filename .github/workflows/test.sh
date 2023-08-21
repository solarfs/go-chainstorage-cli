#!/bin/sh
set -ex


token="github_api_token"

tag=$(git describe --tags `git rev-list --tags --max-count=1`)

echo $tag
