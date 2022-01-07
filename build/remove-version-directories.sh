#!/bin/sh

# check
/usr/bin/find . -type d -name "vs-*" 

# execute
/usr/bin/find . -type d -name "vs-*" -exec rm -rf {} \;