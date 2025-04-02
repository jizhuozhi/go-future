package future

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func getIP() (net.IP, error) {
	time.Sleep(time.Second / 2)
	return net.IPv4(1, 1, 1, 1), nil
}

func getUser() (string, error) {
	time.Sleep(time.Second / 10)
	return "Jack", nil
}

func getUserCountry(info Tuple2[string, net.IP], err error) (string, error) {
	if err != nil {
		return "", err
	} else if info.Val0 != "Jack" {
		return "", fmt.Errorf("not jack")
	} else if info.Val1.String() != "1.1.1.1" {
		return "", fmt.Errorf("not ip 1.1.1.1")
	}
	time.Sleep(time.Second / 2)
	return "UK", nil
}

func TestOf(t *testing.T) {
	start := time.Now()
	ip := Async(getIP)
	user := Async(getUser)
	country := Then(Of2(user, ip), getUserCountry)
	if country.GetOrDefault("") != "UK" {
		t.Fatalf("unexpected result: UK")
	}
	if time.Since(start) > time.Millisecond*1200 {
		t.Fatalf("exec time > 1200ms")
	}
}
