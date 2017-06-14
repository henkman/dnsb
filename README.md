# dnsb
dns block proxy

- build with go build -ldflags="-H windowsgui -s -w"
- create dnsb.block file in same directory as executable and put site names in there - one per line.
- if you put site.org every sub domain of site.org will be blocked.
- if you put sub.site.org only the domain sub and everything under it will be blocked
