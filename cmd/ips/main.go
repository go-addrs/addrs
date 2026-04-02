package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	// TODO allow mixing ipv4 and ipv6 addresses
	"gopkg.in/addrs.v1/ipv4"
)

func main() {
	// TODO maybe a https://pkg.go.dev/golang.org/x/tools/cmd/goyacc ?

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "127.0.0.1:53")
		},
	}

	arg1 := os.Args[1]

	ipStr := arg1
	op := "None"
	if strings.Contains(ipStr, "|") {
		slice := strings.SplitN(ipStr, "|", 2)
		ipStr = slice[0]
		op = slice[1]
		if op != "ip4" && op != "ssh_pattern" {
			log.Fatalf("expect op to be 'ip4' or 'ssh_pattern' but found '%s'", op)
		}
	}

	ip, err := ipv4.AddressFromString(ipStr)
	if err != nil {
		if op == "ip4" {
			addrs, err := r.LookupIP(context.Background(), "ip4", ipStr)
			if err != nil {
				log.Fatalf("failed to lookup '%s': %v", ipStr, err)
			}
			log.Printf("found: '%s' '%v'", ipStr, addrs)
			if len(addrs) == 0 {
				log.Fatalf("'%s' not found", ipStr)
			}
			// TODO maybe return all results as a set and have another op to select one
			ip, err = ipv4.AddressFromNetIP(addrs[0])
			if err != nil {
				log.Fatalf("failed to lookup '%s'", ipStr)
			}
		}
		if op == "ssh_pattern" {
			prefix, err := ipv4.PrefixFromString(ipStr)
			if err != nil {
				log.Fatalf("failed to parse prefix for ssh_pattern '%s': %v", ipStr, err)
			}
			prefix.Set().WalkPrefixes(func(prefix ipv4.Prefix) bool {
				// TODO be sure that 0.0.0.0/0 works
				for i := 0; i < 32; i += 8 {
					if prefix.Length() <= i {
						prefixes := []string{}
						prefix.Set().Build(func(s ipv4.Set_) bool {
							for !s.IsEmpty() {
								p, _ := s.Set().FindAvailablePrefix(ipv4.Set{}, uint32(i))
								s.Remove(p)

								pStr := p.Network().Address().String()
								n := (32 - i) / 8
								pStr = pStr[:len(pStr)-2*n]
								pStr += strings.Repeat(".*", n)
								prefixes = append(prefixes, pStr)
							}
							return true
						})
						fmt.Println(strings.Join(prefixes, ","))
						os.Exit(0)
					}
				}
				return true
			})
			os.Exit(1)
		}
	}

	arg2 := os.Args[2]
	arg3 := os.Args[3]

	// TODO the only thing that I support now, hehe
	if arg2 != "in" {
		log.Fatalf("keyword 'in' expected as second argument but found '%s' instead", arg2)
	}
	set, err := ipv4.PrefixFromString(arg3)
	if err != nil {
		log.Fatalf("failed to parse '%s' as an IP set", arg3)
	}
	if set.Contains(ip) {
		os.Exit(0)
	}
	os.Exit(1)
}
