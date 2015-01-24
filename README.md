tcpdumplocator
==============
Simple program that parses tcpdump output and prints location data for packets.

It watches for addresses that send a lot of data, and when enough packets have
been received in a short enough amount of time, prints the location for each
address.

Usage:
------
```
  -g="/var/db/GeoLite/GeoLite2-City.mmdb": GeoLite City database.  May be downloaded from http://dev.maxmind.com/geoip/geoip2/geolite2/.  If missing, no geolocation data will be displayed.
  -p=32: Print a line after this many packets.
  -t=2s: Reset count to 0 if no packets have been seen in this long.  Other suffixes, such as m and ms may be used.
  -x="127.*,255.255.255.0,192.168.*,10\\..*": Comma-separated list of regular expressions describing IP addresses to ignore.  They will be surrounded with '^' and '$' to prevent accidental matches.
```

Geolocation
-----------
Location lookups are done with the MaxMind GeoLite2 City database.  It can be
downloaded from http://dev.maxmind.com/geoip/geoip2/geolite2/

```
Example Usage:

tcpdump -lnni em0 'udp' | ./tcpdumplocator
```
