// This file is part of arduino-serial-utils
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package serialutils

import (
	"fmt"

	"go.bug.st/serial"
)

// PortsMapper is a function that returns a map of available serial ports.
type PortsMapper func() (map[string]bool, error)

// DefaultPortMapper returns a PortsMapper that lists the available serial ports
// using the go.bug.st/serial library enumerator.
func DefaultPortMapper() (map[string]bool, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("listing serial ports: %w", err)
	}
	res := map[string]bool{}
	for _, port := range ports {
		res[port] = true
	}
	return res, nil
}
