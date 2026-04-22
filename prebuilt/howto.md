# HOWTO MAKE PREBUILT

Sqinn-go embeds prebuilt sqinn binaries for convenience.
Currently it embeds sqinn for the following platforms:

- linux_amd64
- windows_amd64
- darwin_amd64
- darwin_arm64.

To update them, do this:

Download latest sqinn builds from https://github.com/cvilsmeier/sqinn/releases
into `Downloads/` directory.

~~~
cd ~/Downloads
unzip dist-darwin-amd64.zip   -d dist-darwin-amd64
unzip dist-darwin-arm64.zip   -d dist-darwin-arm64
unzip dist-linux-amd64.zip    -d dist-linux-amd64
unzip dist-windows-amd64.zip  -d dist-windows-amd64
~~~

~~~
cd /path/to/sqinn-go/src/prebuilt
cat ~/Downloads/dist-linux-amd64/sqinn       | gzip > linux-amd64.gz
cat ~/Downloads/dist-windows-amd64/sqinn.exe | gzip > windows-amd64.gz
cat ~/Downloads/dist-darwin-amd64/sqinn      | gzip > darwin-amd64.gz
cat ~/Downloads/dist-darwin-arm64/sqinn      | gzip > darwin-arm64.gz
~~~

That's it.

Each prebuilt sqinn binary was built by a github runner supported by
github.com. If you do not trust Microsoft (or whoever owns github.com at the
moment), you can always build sqinn yourself and use that.

Also, if you need sqinn for a platform that is not officially supported by
github.com, e.g. arm32, you have to build sqinn yourself and use that.

Also, if you need special SQLite features (e.g. mmap size greater than 2G),
you have to build sqinn yourself and use that.

