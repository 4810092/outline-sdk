//go:build ios

package tun2socks_android

import "errors"

type iosTunnel struct{}

func startTunOverSocksPlatform(fd int, host string, port int) (tunnelImpl, error) {
	// Пока iOS-реализация не готова, возвращаем ошибку,
	// но экспортируемый API уже существует, и gomobile сможет собрать xcframework.
	_ = fd
	_ = host
	_ = port
	return &iosTunnel{}, errors.New("tun2socks_android: iOS implementation not ready yet")
}

func (t *iosTunnel) stop() error {
	// Заглушка: ничего не делает.
	return nil
}
