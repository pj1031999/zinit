# zinit
System in zram


### Description
This is wrapper for `/sbin/init` which mount your root filesystem in zram device using overlayfs, then it spawns `/bin/sh`. All changes in this filesystem are written to ram.

In order to run add to kernel cmdline `init=/path/to/zinit`
