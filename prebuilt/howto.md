HOWTO MAKE PREBUILT
--------------------------------------

Sqinn-go embeds prebuilt sqinn binaries for convenience.
Currently it embeds linux_amd64 and windows_amd64.

To update them, do this:

- Download latest sqinn builds from https://github.com/cvilsmeier/sqinn/releases

- Extract dist-*.zip archives (e.g. in Downloads directory)
    
    cd prebuilt
    cat ~/Downloads/dist-linux-amd64/sqinn       | gzip > linux-amd64.gz
    cat ~/Downloads/dist-windows-amd64/sqinn.exe | gzip > windows-amd64.gz

That's it.
