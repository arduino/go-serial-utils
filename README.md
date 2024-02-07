## go-serial-utils: a golang library with a bunch of procedures to handle serial ports on Arduino (and compatible) boards.

### Board reset

To reset a board you must use the `Reset` method.

```go
Reset(portToTouch string, wait bool, dryRun bool, portsMapper PortsMapper, cb *ResetProgressCallbacks) (string, error)
```

`Reset` will reset a board using the 1200 bps port-touch and waits for the bootloader port that is returned.
Both reset and wait are optional:
- if `portToTouch` is the empty string "" the reset will be skipped
- if `wait` is false waiting will be skipped
If `wait` is true, this function will wait for a new port to appear after the reset and returns it. If a new port can not be detected or if the `wait` parameter is `false`, then the empty string is returned.

If `dryRun` is set to `true` this function will only emulate the port reset without actually performing it, this is useful to mockup for unit-testing and CI. In dryRun mode if the `portToTouch` ends with `"999"` and `wait` is `true`, the function will return a new "mocked" bootloader port as `portToTouch+"0"`.

`portMapper` is a method called to obtain the current serial port list. If `portMapper` is `nil` the default internal port mapper will be used.

`cb` is a struct defining a bunch of callback functions called during the reset operation to provide progress feedback to the caller.

## Security

If you think you found a vulnerability or other security-related bug in this project, please read our
[security policy](https://github.com/arduino/go-paths-helper/security/policy) and report the bug to our Security Team 🛡️
Thank you!

e-mail contact: security@arduino.cc
