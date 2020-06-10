#!/bin/sh

echo HOME is $HOME
echo we are in `pwd`
echo installing sqinn
curl -L https://github.com/cvilsmeier/sqinn/releases/download/v1.0.0/sqinn-dist-1.0.0.tar.gz -o /tmp/sqinn-dist-1.0.0.tar.gz
tar -C /tmp -xf /tmp/sqinn-dist-1.0.0.tar.gz
chmod a+x /tmp/sqinn-dist-1.0.0/linux_amd64/sqinn
echo ls -al
ls -al
echo ls -alR /tmp/sqinn-dist-1.0.0/
ls -alR /tmp/sqinn-dist-1.0.0
echo ls -alR /tmp/sqinn-dist-1.0.0/linux_amd64
ls -alR /tmp/sqinn-dist-1.0.0/linux_amd64

