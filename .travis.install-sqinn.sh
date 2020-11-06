#!/bin/sh

echo install sqinn
curl -L https://github.com/cvilsmeier/sqinn/releases/download/v1.1.2/sqinn-dist-1.1.2.tar.gz -o /tmp/sqinn-dist-1.1.2.tar.gz
tar -C /tmp -xf /tmp/sqinn-dist-1.1.2.tar.gz
chmod a+x /tmp/sqinn-dist-1.1.2/linux_amd64/sqinn

