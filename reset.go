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
	"runtime"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Touch1200bps open and close the serial port at 1200 bps. This is used
// on many Arduino (and compatible) boards as a signal to put the MCU
// in bootloader mode.
func Touch1200bps(port string) error {
	p, err := serial.Open(port, &serial.Mode{BaudRate: 1200})
	if err != nil {
		return fmt.Errorf("opening port at 1200bps: %w", err)
	}

	if runtime.GOOS != "windows" {
		// This is not required on Windows
		// TODO: Investigate if it can be removed for other OS too

		// Set DTR to false
		if err = p.SetDTR(false); err != nil {
			_ = p.Close()
			return fmt.Errorf("setting DTR to OFF: %w", err)
		}
	}

	// Close serial port
	_ = p.Close()

	// Scanning for available ports seems to open the port or
	// otherwise assert DTR, which would cancel the WDT reset if
	// it happens within 250 ms. So we wait until the reset should
	// have already occurred before going on.
	time.Sleep(500 * time.Millisecond)

	return nil
}

// ResetProgressCallbacks is a struct that defines a bunch of function callback
// to observe the Reset function progress.
type ResetProgressCallbacks struct {
	// TouchingPort is called to signal the 1200-bps touch of the reported port
	TouchingPort func(port string)
	// WaitingForNewSerial is called to signal that we are waiting for a new port
	WaitingForNewSerial func()
	// BootloaderPortFound is called to signal that the wait is completed and to
	// report the port found, or the empty string if no ports have been found and
	// the wait has timed-out.
	BootloaderPortFound func(port string)
	// Debug reports messages useful for debugging purposes. In normal conditions
	// these messages should not be displayed to the user.
	Debug func(msg string)
}

// Reset will reset a board using the 1200 bps port-touch and waits for the bootloader port that is returned.
// Both reset and wait are optional:
// - if `portToTouch` is the empty string "" the reset will be skipped
// - if `wait` is false waiting will be skipped
// If `wait` is true, this function will wait for a new port to appear after the reset and returns it. If
// a new port can not be detected or if the `wait` parameter is `false`, then the empty string is returned.
//
// If `dryRun` is set to `true` this function will only emulate the port reset without actually performing
// it, this is useful to mockup for unit-testing and CI. In dryRun mode if the `portToTouch` ends with
// `"999"` and `wait` is `true`, the function will return a new "mocked" bootloader port as `portToTouch+"0"`.
//
// `portMapper` is a method called to obtain the current serial port list. If `portMapper` is `nil` the
// default internal port mapper will be used.
//
// `cb` is a struct defining a bunch of callback functions called during the reset operation to provide
// progress feedback to the caller.
func Reset(portToTouch string, wait bool, dryRun bool, portsMapper PortsMapper, cb *ResetProgressCallbacks) (string, error) {
	if portsMapper == nil {
		portsMapper = DefaultPortMapper // non dry-run default
	}
	if dryRun {
		emulatedPort := portToTouch
		portsMapper = func() (map[string]bool, error) {
			res := map[string]bool{}
			if emulatedPort != "" {
				res[emulatedPort] = true
			}
			if strings.HasSuffix(emulatedPort, "999") {
				emulatedPort += "0"
			} else if emulatedPort == "" {
				emulatedPort = "newport"
			}
			return res, nil
		}
	}

	last, err := portsMapper()
	if cb != nil && cb.Debug != nil {
		cb.Debug(fmt.Sprintf("LAST: %v", last))
	}
	if err != nil {
		return "", err
	}

	if portToTouch != "" && last[portToTouch] {
		if cb != nil && cb.Debug != nil {
			cb.Debug(fmt.Sprintf("TOUCH: %v", portToTouch))
		}
		if cb != nil && cb.TouchingPort != nil {
			cb.TouchingPort(portToTouch)
		}
		if dryRun {
			// do nothing!
		} else {
			if err := Touch1200bps(portToTouch); err != nil && !wait {
				return "", fmt.Errorf("1200-bps touch: %w", err)
			}
		}
	}

	if !wait {
		return "", nil
	}
	if cb != nil && cb.WaitingForNewSerial != nil {
		cb.WaitingForNewSerial()
	}

	deadline := time.Now().Add(10 * time.Second)
	if dryRun {
		// use a much lower timeout in dryRun
		deadline = time.Now().Add(100 * time.Millisecond)
	}
	for time.Now().Before(deadline) {
		now, err := portsMapper()
		if err != nil {
			return "", err
		}
		if cb != nil && cb.Debug != nil {
			cb.Debug(fmt.Sprintf("WAIT: %v", now))
		}
		hasNewPorts := false
		for p := range now {
			if !last[p] {
				hasNewPorts = true
				break
			}
		}

		if hasNewPorts {
			if cb != nil && cb.Debug != nil {
				cb.Debug("New ports found!")
			}

			// on OS X, if the port is opened too quickly after it is detected,
			// a "Resource busy" error occurs, add a delay to workaround.
			// This apply to other platforms as well.
			time.Sleep(time.Second)

			// Some boards have a glitch in the bootloader: some user experienced
			// the USB serial port appearing and disappearing rapidly before
			// settling.
			// This check ensure that the port is stable after one second.
			check, err := portsMapper()
			if err != nil {
				return "", err
			}
			if cb != nil && cb.Debug != nil {
				cb.Debug(fmt.Sprintf("CHECK: %v", check))
			}
			for p := range check {
				if !last[p] {
					if cb != nil && cb.BootloaderPortFound != nil {
						cb.BootloaderPortFound(p)
					}
					return p, nil // Found it!
				}
			}
			if cb != nil && cb.Debug != nil {
				cb.Debug("Port check failed... still waiting")
			}
		}

		last = now
		time.Sleep(250 * time.Millisecond)
	}

	if cb != nil && cb.BootloaderPortFound != nil {
		cb.BootloaderPortFound("")
	}
	return "", nil
}
