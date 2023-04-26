package vsock

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/mdlayher/vsock"
)

func SchemeMatches(addr string) bool {
	return strings.HasPrefix(addr, "vsock://")
}

func parse(addr string) (uint32, uint32, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return 0, 0, fmt.Errorf("bad vsock address %q: %w", addr, err)
	}

	c, err := strconv.ParseUint(u.Hostname(), 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid vsock cid in %q: %w", addr, err)
	}

	p, err := strconv.ParseUint(u.Port(), 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid vsock port in %q: %w", addr, err)
	}

	return uint32(c), uint32(p), nil
}

func Listen(addr string) (net.Listener, error) {
	_, p, err := parse(addr)
	if err != nil {
		return nil, fmt.Errorf("vsock listen failed for %q: %w", addr, err)
	}

	l, err := vsock.Listen(p, nil)
	if err != nil {
		return nil, fmt.Errorf("vsock listen failed for %q: %w", addr, err)
	}

	return l, nil
}

func Dial(ctx context.Context, addr string) (net.Conn, error) {
	cid, port, err := parse(addr)
	if err != nil {
		return nil, fmt.Errorf("vsock dial failed for %q: %w", addr, err)
	}

	// TODO(tho) handle ctx

	return vsock.Dial(cid, port, nil)
}

func Dialer(ctx context.Context, addr string) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) { return Dial(ctx, addr) }
}
