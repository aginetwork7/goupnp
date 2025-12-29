package httpu

import (
	"log/slog"
	"net"
	"runtime"
	"strings"
	"syscall"
)

func setMulticastInterface(conn any, ipAddr string) {
	cc, ok := conn.(*net.UDPConn)
	if !ok {
		slog.Warn("setMulticastInterface: connection is not a *net.UDPConn, skipping multicast setup")
		return
	}

	rawConn, err := cc.SyscallConn()
	if err != nil {
		slog.Error("Failed to get raw connection", "error", err, "ipAddr", ipAddr)
		return
	}

	if rawConn == nil {
		return
	}

	if !strings.Contains(ipAddr, ":") {
		return
	}

	err = rawConn.Control(func(fd uintptr) {
		host := ipAddr[:strings.LastIndex(ipAddr, ":")]

		switch runtime.GOOS {
		case "darwin", "linux":
			localIP := net.ParseIP(host).To4()
			if len(localIP) == 4 {
				addr := [4]byte{localIP[0], localIP[1], localIP[2], localIP[3]}
				err1 := syscall.SetsockoptInet4Addr(int(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_IF, addr)
				if err1 != nil {
					slog.Error("Failed to set socket option", "error", err1, "ipAddr", ipAddr)
					return
				}
			} else {
				slog.Warn("Unsupported IP address family for multicast on %q: only IPv4 is supported", ipAddr)
			}
		default:
			slog.Error("Unsupported OS", "os", runtime.GOOS)
		}
	})
	if err != nil {
		slog.Error("Failed to set socket option", "error", err, "ipAddr", ipAddr)
	}
}
