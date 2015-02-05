/*
 * tcpdumplocator.go
 * Program to print out locations of packets.  Meant to parse tcpdump output..
 * by J. Stuart McMurray
 * created 20150123
 * last modified 20150123
 *
 * Copyright (c) 2015 J. Stuart McMurray <kd5pbo@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	dbfile = flag.String("g", "/var/db/GeoLite/GeoLite2-City.mmdb",
		"GeoLite City database.  May be downloaded from "+
			"http://dev.maxmind.com/geoip/geoip2/geolite2/.  If "+
			"missing, no geolocation data will be displayed.")
	printAfter = flag.Int("p", 32, "Print a line after this many packets.")
	ignore     = flag.String("x", `127.*,255.255.255.0,192.168.*,10\..*`,
		"Comma-separated list of regular expressions describing IP "+
			"addresses to ignore.  They will be surrounded with "+
			"'^' and '$' to prevent accidental matches.")
	timeoff = flag.Duration("t", 2*time.Second, "Reset count to 0 if no "+
		"packets have been seen in this long.  Other suffixes, such "+
		"as m and ms may be used.")
	/* IP addres regex */
	ADDRRE = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	/* GeoLite City Database */
	geodb *geoip2.Reader
	/* Collection of seen IP addresses */
	seenIPs = map[string]seenIP{}
	/* Holds the ip-ignoring regexen */
	ignoreRegex = []*regexp.Regexp{}
	/* Last-printed IP address */
	lastPrinted = ""
)

/* An IP address we've seen */
type seenIP struct {
	addr     string
	nSeen    int
	lastSeen time.Time
}

/* All the IP addresses we've seen */

func main() { os.Exit(mymain()) }
func mymain() int {
	flag.Parse()
	/* Parse regular expressions */
	for _, ex := range strings.Split(*ignore, ",") {
		r, err := regexp.Compile("^" + ex + "$")
		if nil != err {
			fmt.Fprintf(os.Stderr, "Unable to compile %v: %v\n",
				ex, r)
			return -1
		}
		ignoreRegex = append(ignoreRegex, r)
	}
	/* Attempt to open geolite database */
	var err error
	geodb, err = geoip2.Open(*dbfile)
	if nil != err {
		fmt.Printf("Could not open %v: %v\n", *dbfile, err)
	}
	/* Read from stdin */
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		l := scanner.Text()
		/* Dig addresses out of the line */
		addrs := ADDRRE.FindAllString(l, -1)
		/* Don't bother if there's no addresses */
		if 0 == len(addrs) {
			continue
		}
		/* Handle the seen addresses */
	ADDRS:
		for _, a := range addrs {
			/* Ignore excluded addresses */
			for _, r := range ignoreRegex {
				if r.MatchString(a) {
					continue ADDRS
				}
			}
			seen_addr(a)
		}
	}
	if err := scanner.Err(); err != nil {
		/* EOF is normal */
		if io.EOF == err {
			return 0
		}
		/* Otherwise, die */
		fmt.Printf("Error: %v", err)
		return -1
	}
	return 0
}

/* Handle addresses we've seen */
func seen_addr(a string) {
	/* Pull the record from seenIPs, if it's there */
	r, ok := seenIPs[a]
	/* If it's new, save it */
	if !ok {
		r = seenIP{addr: a, nSeen: 0, lastSeen: time.Now()}
	}
	r.nSeen++
	/* Reset the count if it's too old */
	if *timeoff < time.Now().Sub(r.lastSeen) {
		r.nSeen = 0
	}
	r.lastSeen = time.Now()
	/* Update the collection */
	seenIPs[a] = r
	/* Print output after printAfrter and every printEvery */
	if *printAfter == r.nSeen {
		print_addr(a)
	}
}

/* Print an Address with geo info */
func print_addr(a string) {
	/* Don't print it if it's the last to be printed */
	if a == lastPrinted {
		return
	}
	/* Get GeoIP information */
	geoinfo := ""
	ip := net.ParseIP(a)
	record, err := geodb.City(ip)
	if err != nil {
		/* Can't geoip it */
		geoinfo = err.Error()
	} else if record.Traits.IsAnonymousProxy {
		/* Tor ? */
		geoinfo = "Anonymous Proxy"
	} else if "" == record.Country.IsoCode {
		/* Internal ? */
		geoinfo = "Countryless"
	} else {
		/* Subdivisions may not happen */
		s := ""
		subdivision := record.Subdivisions
		if 0 == len(subdivision) {
			s = "No Subdivision (State)"
		} else {
			for _, v := range subdivision {
				s += getEN(v.Names)
			}
		}
		/* Put the geo info together in a string */
		geoinfo = fmt.Sprintf("%v, %v, %v, %v",
			record.Country.IsoCode,
			getEN(record.Country.Names),
			s,
			getEN(record.City.Names))
	}

	/* Move cursor down to the beginning of the next row, clear the
	row, print the address and count */
	log.Printf("%15v  %v\n", a, geoinfo)
	lastPrinted = a
}

/* Get the EN (or other lang) value in the map, and failing that something
else */
func getEN(m map[string]string) string {
	/* Try the EN value */
	for _, l := range []string{"en", "es", "fr", "de"} {
		/* Clean up the language */
		l = strings.ToLower(strings.TrimSpace(l))
		/* Skip it if it's empty (like for en,,es) */
		if 0 == len(l) {
			continue
		}
		/* Try getting the name for the language */
		v, ok := m[strings.TrimSpace(strings.ToLower(l))]
		/* If we got one, return it */
		if ok {
			return v
		}
	}
	/* If we didn't get anything, give the list of allowable names */
	name := ""
	ks := ""
	/* EN failed, get ALL the languages (and the longest value) */
	for k, v := range m {
		ks += "|" + k
		if len(v) > len(name) {
			name = v
		}
	}
	if 0 == len(ks) {
		return "Missing"
	}
	/* Put it all together */
	for '|' == ks[0] {
		ks = ks[1:len(ks)]
	}
	return fmt.Sprintf("[%v]%v", ks, name)
}
