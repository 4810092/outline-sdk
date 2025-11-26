//go:build android

package tun2socks_android

import (
	"context"
	"fmt"
	"time"
	// TODO: сюда позже будут добавлены реальные импорты outline-sdk,
	// например:
	// "github.com/Jigsaw-Code/outline-sdk/network"
	// "github.com/Jigsaw-Code/outline-sdk/transport/socks5"
)

type androidTunnel struct {
	cancel context.CancelFunc
}

func startTunOverSocksPlatform(fd int, host string, port int) (tunnelImpl, error) {
	// TODO: заменить на реальную реализацию tun2socks с использованием outline-sdk:
	//  1. Обернуть TUN-устройство по fd.
	//  2. Создать SOCKS5-transport к host:port.
	//  3. Запустить цикл tun2socks в отдельной goroutine.

	ctx, cancel := context.WithCancel(context.Background())

	// Временная заглушка: просто логируем параметры и имитируем работу.
	_ = fd
	_ = host
	_ = port

	go func() {
		// Заглушка "работы" — чтобы было на что повесить cancel.
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Здесь позже будет логика обработки пакетов.
				_ = fmt.Sprintf("tun2socks_android tick fd=%d host=%s port=%d", fd, host, port)
			}
		}
	}()

	return &androidTunnel{cancel: cancel}, nil
}

func (t *androidTunnel) stop() error {
	if t == nil {
		return nil
	}
	if t.cancel != nil {
		t.cancel()
	}
	return nil
}
