#!/bin/sh

git --version > gitversion

git config --global user.email "1071318859@qq.com"

git config --global user.name "1071318859"

git add .

git commit -m "update" -a

git push
