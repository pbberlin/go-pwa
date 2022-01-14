#!/bin/sh

cd ../app-bucket

# check
/usr/bin/find . -type d -name "vs-*" 

# execute
/usr/bin/find . -type d -name "vs-*" -exec rm -rf {} \;