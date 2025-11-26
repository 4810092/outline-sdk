package tun2socks_android

// TunnelHandle — дескриптор запущенного туннеля TUN↔SOCKS.
// Его методы будут доступны из Kotlin/Swift.
type TunnelHandle struct {
	impl tunnelImpl
}

// tunnelImpl — внутренняя реализация под каждую платформу (android/ios).
type tunnelImpl interface {
	stop() error
}

// StartTunOverSocks запускает мост TUN→SOCKS на данном файловом дескрипторе.
// fd — файловый дескриптор TUN-интерфейса (из VpnService.Builder.establish()).
// host, port — адрес локального SOCKS-прокси (например "127.0.0.1", 10808).
func StartTunOverSocks(fd int, host string, port int) (*TunnelHandle, error) {
	impl, err := startTunOverSocksPlatform(fd, host, port)
	if err != nil {
		return nil, err
	}
	return &TunnelHandle{impl: impl}, nil
}

// Stop останавливает туннель и освобождает ресурсы.
func (h *TunnelHandle) Stop() error {
	if h == nil || h.impl == nil {
		return nil
	}
	return h.impl.stop()
}
