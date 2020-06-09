#!/bin/sh

echo HOME is $HOME
echo we are in `pwd`
echo installing sqinn
curl -L https://github.com/cvilsmeier/sqinn/releases/download/1.0.0/sqinn-dist-1.0.0.tar.gz -o /tmp/sqinn-dist-1.0.0.tar.gz
tar -C /tmp -xf /tmp/sqinn-dist-1.0.0.tar.gz
echo ll 
ls -al
echo ll /tmp
ls -al /tmp
export SQINN_PATH=/tmp/sqinn-dist-1.0.0/linux_amd64/sqinn
echo SQINN_PATH is $SQINN_PATH

