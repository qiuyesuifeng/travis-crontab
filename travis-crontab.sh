#!/bin/bash
# show an example.

source ~/.bashrc > /dev/null
source ~/.zshrc > /dev/null

go get github.com/qiuyesuifeng/travis-crontab
cd $GOPATH/src/github.com/qiuyesuifeng/travis-crontab
go build -o travis-crontab

./travis-crontab -t $1 -r $2 -b $3
