#!/bin/sh

git --version > gitversion

git add .

git commit -m "update" -a

git push
