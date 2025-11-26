//go:build !android && !ios

package tun2socks_android

import "errors"

type unsupportedTunnel struct{}

func startTunOverSocksPlatform(fd int, host string, port int) (tunnelImpl, error) {
	_ = fd
	_ = host
	_ = port
	return &unsupportedTunnel{}, errors.New("tun2socks_android: unsupported platform")
}

func (t *unsupportedTunnel) stop() error {
	return nil
}
