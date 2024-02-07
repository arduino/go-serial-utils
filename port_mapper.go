// This file is part of arduino-serial-utils
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

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
