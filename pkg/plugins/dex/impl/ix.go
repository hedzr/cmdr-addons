// Copyright © 2020 Hedzr Yeh.

package impl

type (
	// System represents the service manager that is available.
	System interface {
		// String returns a description of the system.
		String() string

		// Detect returns true if the system is available to use.
		Detect() bool

		// Interactive returns false if running under the system service manager
		// and true otherwise.
		Interactive() bool

		// New creates a new service for this system.
		New(i Operate, c *Config) (Service, error)
	}

	// Interface represents the service interface for a program. Start runs before
	// the hosting process is granted control and Stop runs when control is returned.
	//
	//   1. OS service manager executes user program.
	//   2. User program sees it is executed from a service manager (IsInteractive is false).
	//   3. User program calls Service.Run() which blocks.
	//   4. Interface.Start() is called and quickly returns.
	//   5. User program runs.
	//   6. OS service manager signals the user program to stop.
	//   7. Interface.Stop() is called and quickly returns.
	//      - For a successful exit, os.Exit should not be called in Interface.Stop().
	//   8. Service.Run returns.
	//   9. User program should quickly exit.
	Operate interface {
		// Start provides a place to initiate the service. The service doesn't not
		// signal a completed start until after this function returns, so the
		// Start function must not take more then a few seconds at most.
		Start(s Service) error

		// Stop provides a place to clean up program execution before it is terminated.
		// It should not take more then a few seconds to execute.
		// Stop should not call os.Exit directly in the function.
		Stop(s Service) error
	}

	// Service represents a service that can be run or controlled.
	Service interface {
		// Run should be called shortly after the program entry point.
		// After Interface.Stop has finished running, Run will stop blocking.
		// After Run stops blocking, the program must exit shortly after.
		Run() error

		// Start signals to the OS service manager the given service should start.
		Start() error

		// Stop signals to the OS service manager the given service should stop.
		Stop() error

		// Restart signals to the OS service manager the given service should stop then start.
		Restart() error

		// Install setups up the given service in the OS service manager. This may require
		// greater rights. Will return an error if it is already installed.
		Install() error

		// Uninstall removes the given service from the OS service manager. This may require
		// greater rights. Will return an error if the service is not present.
		Uninstall() error

		// Opens and returns a system logger. If the user program is running
		// interactively rather then as a service, the returned logger will write to
		// os.Stderr. If errs is non-nil errors will be sent on errs as well as
		// returned from Logger's functions.
		Logger(errs chan<- error) (Logger, error)

		// SystemLogger opens and returns a system logger. If errs is non-nil errors
		// will be sent on errs as well as returned from Logger's functions.
		SystemLogger(errs chan<- error) (Logger, error)

		// String displays the name of the service. The display name if present,
		// otherwise the name.
		String() string

		// Platform displays the name of the system that manages the service.
		// In most cases this will be the same as service.Platform().
		Platform() string

		// Status returns the current service status.
		Status() (Status, error)
	}

	// Logger writes to the system log.
	Logger interface {
		Error(v ...interface{}) error
		Warning(v ...interface{}) error
		Info(v ...interface{}) error

		Errorf(format string, a ...interface{}) error
		Warningf(format string, a ...interface{}) error
		Infof(format string, a ...interface{}) error
	}

	// Config provides the setup for a Service. The Name field is required.
	Config struct {
		Name        string   // Required name of the service. No spaces suggested.
		DisplayName string   // Display name, spaces allowed.
		Description string   // Long description of service.
		UserName    string   // Run as username.
		Arguments   []string // Run with arguments.

		// Optional field to specify the executable for service.
		// If empty the current executable is used.
		Executable string

		// Array of service dependencies.
		// Not yet implemented on Linux or OS X.
		Dependencies []string

		// The following fields are not supported on Windows.
		WorkingDirectory string // Initial working directory.
		ChRoot           string

		// System specific options.
		//  * OS X
		//    - LaunchdConfig string ()      - Use custom launchd config
		//    - KeepAlive     bool   (true)
		//    - RunAtLoad     bool   (false)
		//    - UserService   bool   (false) - Install as a current user service.
		//    - SessionCreate bool   (false) - Create a full user session.
		//  * POSIX
		//    - SystemdScript string ()                 - Use custom systemd script
		//    - UpstartScript string ()                 - Use custom upstart script
		//    - SysvScript    string ()                 - Use custom sysv script
		//    - RunWait       func() (wait for SIGNAL)  - Do not install signal but wait for this function to return.
		//    - ReloadSignal  string () [USR1, ...]     - Signal to send on reaload.
		//    - PIDFile       string () [/run/prog.pid] - Location of the PID file.
		//    - LogOutput     bool   (false)            - Redirect StdErr & StdOut to files.
		Option KeyValue
	}
)

// Status represents service status as an byte value
type Status byte

// Status of service represented as an byte
const (
	StatusUnknown Status = iota // Status is unable to be determined due to an error or it was not installed.
	StatusRunning
	StatusStopped
)

// KeyValue provides a list of platform specific options. See platform docs for
// more details.
type KeyValue map[string]interface{}

// bool returns the value of the given name, assuming the value is a boolean.
// If the value isn't found or is not of the type, the defaultValue is returned.
func (kv KeyValue) bool(name string, defaultValue bool) bool {
	if v, found := kv[name]; found {
		if castValue, is := v.(bool); is {
			return castValue
		}
	}
	return defaultValue
}

// int returns the value of the given name, assuming the value is an int.
// If the value isn't found or is not of the type, the defaultValue is returned.
func (kv KeyValue) int(name string, defaultValue int) int {
	if v, found := kv[name]; found {
		if castValue, is := v.(int); is {
			return castValue
		}
	}
	return defaultValue
}

// string returns the value of the given name, assuming the value is a string.
// If the value isn't found or is not of the type, the defaultValue is returned.
func (kv KeyValue) string(name string, defaultValue string) string {
	if v, found := kv[name]; found {
		if castValue, is := v.(string); is {
			return castValue
		}
	}
	return defaultValue
}

// float64 returns the value of the given name, assuming the value is a float64.
// If the value isn't found or is not of the type, the defaultValue is returned.
func (kv KeyValue) float64(name string, defaultValue float64) float64 {
	if v, found := kv[name]; found {
		if castValue, is := v.(float64); is {
			return castValue
		}
	}
	return defaultValue
}

// funcSingle returns the value of the given name, assuming the value is a float64.
// If the value isn't found or is not of the type, the defaultValue is returned.
func (kv KeyValue) funcSingle(name string, defaultValue func()) func() {
	if v, found := kv[name]; found {
		if castValue, is := v.(func()); is {
			return castValue
		}
	}
	return defaultValue
}
