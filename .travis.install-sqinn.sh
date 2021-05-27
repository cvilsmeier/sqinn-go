#!/bin/sh

echo install sqinn
curl -L https://github.com/cvilsmeier/sqinn/releases/download/v1.1.8/sqinn-dist-1.1.8.tar.gz -o /tmp/sqinn-dist-1.1.8.tar.gz
tar -C /tmp -xf /tmp/sqinn-dist-1.1.8.tar.gz
chmod a+x /tmp/sqinn-dist-1.1.8/linux_amd64/sqinn

