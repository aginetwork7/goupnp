package httpu

import (
	"log/slog"
	"net"
	"runtime"
	"syscall"
)

func joinMultiCast(conn any, ipAddr string) {
	cc, ok := conn.(*net.UDPConn)
	if !ok {
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

	err = rawConn.Control(func(fd uintptr) {
		switch runtime.GOOS {
		case "darwin", "linux":
			localIP := net.ParseIP(ipAddr).To4()
			if len(localIP) == 4 {
				addr := [4]byte{localIP[0], localIP[1], localIP[2], localIP[3]}
				err = syscall.SetsockoptInet4Addr(int(fd), syscall.IPPROTO_IP, syscall.IP_MULTICAST_IF, addr)
				if err != nil {
					slog.Error("Failed to set socket option", "error", err, "ipAddr", ipAddr)
					return
				}
			}
		default:
			slog.Error("Unsupported OS", "os", runtime.GOOS)
		}
	})
	if err != nil {
		slog.Error("Failed to set socket option", "error", err, "ipAddr", ipAddr)
	}
}
