#!/bin/bash
go build
go install
gox -osarch="linux/amd64" --output=/Users/mihuan/www/mihuan-new/productServer

sshpass -p liuhanzeng scp /Users/mihuan/www/mihuan-new/productServer liuhanzeng@172.16.1.44:devspace/mihuan-new/productServer
