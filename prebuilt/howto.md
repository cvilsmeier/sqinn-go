HOWTO MAKE PREBUILT
--------------------------------------

Sqinn-go embeds prebuilt sqinn binaries for convenience.
Currently it embeds linux_amd64 and windows_amd64.

To update them, do this:

- Download latest sqinn builds from https://github.com/cvilsmeier/sqinn/releases

- Extract `dist-*.zip` archives (e.g. in `Downloads/` directory)

~~~    
cd prebuilt
cat ~/Downloads/dist-linux-amd64/sqinn       | gzip > linux-amd64.gz
cat ~/Downloads/dist-windows-amd64/sqinn.exe | gzip > windows-amd64.gz
cat ~/Downloads/dist-darwin-amd64/sqinn      | gzip > darwin-amd64.gz
cat ~/Downloads/dist-darwin-arm64/sqinn      | gzip > darwin-arm64.gz
~~~

That's it.

Each prebuilt sqinn binary was built by a github runner supported by github.com.
If you do not trust github.com (or whatever company owns github.com at the moment),
you can always build sqinn yourself and use that.

Also, if you need sqinn for a platform that is not officially supported by github.com,
e.g. arm32, you have to build sqinn yourself and use that.
