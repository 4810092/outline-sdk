// Copyright 2024 The Outline Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tun2socks_android

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/Jigsaw-Code/outline-sdk/network"
	"github.com/Jigsaw-Code/outline-sdk/network/lwip2transport"
	"github.com/Jigsaw-Code/outline-sdk/transport"
	"github.com/Jigsaw-Code/outline-sdk/transport/socks5"
)

// TunnelHandle represents a running tun2socks instance.
type TunnelHandle struct {
	cancel context.CancelFunc
	done   chan struct{}

	tunDevice  network.IPDevice
	lwipDevice network.IPDevice

	stopOnce sync.Once
}

// StartTunOverSocks starts piping traffic between the TUN file descriptor and a SOCKS5 proxy.
func StartTunOverSocks(fd int, socksHost string, socksPort int) (*TunnelHandle, error) {
	if fd <= 0 {
		return nil, fmt.Errorf("invalid tun file descriptor: %d", fd)
	}
	if socksHost == "" {
		return nil, fmt.Errorf("socksHost must not be empty")
	}
	if socksPort <= 0 {
		return nil, fmt.Errorf("invalid socksPort: %d", socksPort)
	}

	address := net.JoinHostPort(socksHost, strconv.Itoa(socksPort))
	endpoint := &transport.TCPEndpoint{
		Dialer:  net.Dialer{},
		Address: address,
	}
	socksClient, err := socks5.NewClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 client: %w", err)
	}
	socksClient.EnablePacket(&transport.UDPDialer{})

	packetProxy, err := network.NewPacketProxyFromPacketListener(socksClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP proxy: %w", err)
	}

	lwipDevice, err := lwip2transport.ConfigureDevice(socksClient, packetProxy)
	if err != nil {
		return nil, fmt.Errorf("failed to configure lwIP device: %w", err)
	}

	tunDevice, err := newTUNDevice(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	handle := &TunnelHandle{
		cancel:     cancel,
		done:       make(chan struct{}),
		tunDevice:  tunDevice,
		lwipDevice: lwipDevice,
	}

	go handle.run(ctx)

	return handle, nil
}

func (h *TunnelHandle) run(ctx context.Context) {
	defer close(h.done)

	var closeOnce sync.Once
	closeDevices := func() {
		closeOnce.Do(func() {
			h.tunDevice.Close()
			h.lwipDevice.Close()
		})
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Traffic from the Android TUN to the lwIP stack.
	go func() {
		defer wg.Done()
		_, _ = io.Copy(h.lwipDevice, h.tunDevice)
		h.cancel()
	}()
	// Traffic from the lwIP stack back to the Android TUN.
	go func() {
		defer wg.Done()
		_, _ = io.Copy(h.tunDevice, h.lwipDevice)
		h.cancel()
	}()

	<-ctx.Done()
	closeDevices()
	wg.Wait()
}

// Stop stops the tunnel and releases resources.
func (h *TunnelHandle) Stop() error {
	if h == nil {
		return nil
	}
	h.stopOnce.Do(func() {
		if h.cancel != nil {
			h.cancel()
		}
		<-h.done
	})
	return nil
}

type tunDevice struct {
	file *os.File
	mtu  int
}

var _ network.IPDevice = (*tunDevice)(nil)

func newTUNDevice(fd int) (network.IPDevice, error) {
	if fd <= 0 {
		return nil, fmt.Errorf("invalid file descriptor: %d", fd)
	}
	file := os.NewFile(uintptr(fd), "tun")
	if file == nil {
		return nil, fmt.Errorf("failed to create file from fd %d", fd)
	}
	return &tunDevice{
		file: file,
		mtu:  1500,
	}, nil
}

func (d *tunDevice) Close() error {
	return d.file.Close()
}

func (d *tunDevice) Read(p []byte) (int, error) {
	return d.file.Read(p)
}

func (d *tunDevice) Write(p []byte) (int, error) {
	if len(p) > d.mtu {
		return 0, network.ErrMsgSize
	}
	return d.file.Write(p)
}

func (d *tunDevice) MTU() int {
	return d.mtu
}
