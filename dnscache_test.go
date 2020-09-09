package dnscache

import (
  "net"
  "sort"
  "testing"
  "time"
)

func TestFetchReturnsAndErrorOnInvalidLookup(t *testing.T) {
  ips, err := New(0).Lookup("invalid.viki.io", true)
  if ips != nil {
    t.Errorf("Expecting nil ips, got %v", ips)
  }
  expected := "lookup invalid.viki.io: no such host"
  if err.Error() != expected {
    t.Errorf("Expecting %q error, got %q", expected, err.Error())
  }
}

func TestFetchReturnsAListOfIps(t *testing.T) {
  e, _ := New(0).Lookup("dnscache.go.test.viki.io", true)
  assertIps(t, e.ips, []string{"1.123.58.13", "31.85.32.110"})
}

func TestCallingLookupAddsTheItemToTheCache(t *testing.T) {
  r := New(0)
  r.Lookup("dnscache.go.test.viki.io", true)
  assertIps(t, r.cache["dnscache.go.test.viki.io"].ips, []string{"1.123.58.13", "31.85.32.110"})
}

func TestFetchLoadsValueFromTheCache(t *testing.T) {
  r := New(0)
  r.cache["invalid.viki.io"] = &ipaccess{ips: []net.IP{net.ParseIP("1.1.2.3")}}
  e, _ := r.Fetch("invalid.viki.io")
  assertIps(t, e.ips, []string{"1.1.2.3"})
}

func TestFetchOneLoadsTheFirstValue(t *testing.T) {
  r := New(0)
  r.cache["something.viki.io"] = &ipaccess{ips: []net.IP{net.ParseIP("1.1.2.3"), net.ParseIP("100.100.102.103")}}
  ip, _ := r.FetchOne("something.viki.io")
  assertIps(t, []net.IP{ip}, []string{"1.1.2.3"})
}

func TestFetchOneStringRoundRobin(t *testing.T) {
  r := New(0)
  r.cache["something.viki.io"] = &ipaccess{ips: []net.IP{net.ParseIP("100.100.102.103"), net.ParseIP("100.100.102.104")}}
  ip1, _ := r.FetchOneString("something.viki.io")
  ip2, _ := r.FetchOneString("something.viki.io")
  if ip1 == ip2 {
    t.Errorf("expected different ips but got %v", ip1)
  }
}

func TestFetchLoadsTheIpAndCachesIt(t *testing.T) {
  r := New(0)
  e, _ := r.Fetch("dnscache.go.test.viki.io")
  assertIps(t, e.ips, []string{"1.123.58.13", "31.85.32.110"})
  assertIps(t, r.cache["dnscache.go.test.viki.io"].ips, []string{"1.123.58.13", "31.85.32.110"})
}

func TestItReloadsTheIpsAtAGivenInterval(t *testing.T) {
  r := New(1)
  r.cache["dnscache.go.test.viki.io"] = nil
  time.Sleep(time.Second * 2)
  if r.cache["dnscache.go.test.viki.io"] != nil {
    t.Errorf("Got an unexpected entry: %v:", r.cache["dnscache.go.test.viki.io"])
  }
}

func assertIps(t *testing.T, actuals []net.IP, expected []string) {
  if len(actuals) != len(expected) {
    t.Errorf("Expecting %d ips, got %d", len(expected), len(actuals))
  }
  sort.Strings(expected)
  for _, ip := range actuals {
    if sort.SearchStrings(expected, ip.String()) == -1 {
      t.Errorf("Got an unexpected ip: %v:", actuals[0])
    }
  }
}