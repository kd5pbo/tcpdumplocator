# tcpdumplocator
Simple program that parses tcpdump output and prints location data for packets.

Requires the MaxMind GeoLite City database:
from http://dev.maxmind.com/geoip/geoip2/geolite2/

This file will have more documentation eventually.

Usage:
```
tcpdump -lnni em0 'udp' | ./tcpdumplocator
```
