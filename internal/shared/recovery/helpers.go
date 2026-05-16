package recovery

import (
	"fmt"
	"log"
)

// SafeGo runs a function in a goroutine with panic recovery
// If a panic occurs, it logs the panic and continues
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				panicInfo := BuildPanicInfo(rec, 3)
				fmt.Print(panicInfo.Format())
			}
		}()
		fn()
	}()
}

// SafeGoWithCallback runs a function in a goroutine with panic recovery
// If a panic occurs, it calls the callback with panic info
func SafeGoWithCallback(fn func(), onPanic func(*PanicInfo)) {
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				panicInfo := BuildPanicInfo(rec, 3)
				fmt.Print(panicInfo.Format())
				if onPanic != nil {
					onPanic(panicInfo)
				}
			}
		}()
		fn()
	}()
}

// SafeCall wraps a function call with panic recovery
// Returns error if panic occurs
func SafeCall(fn func()) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			panicInfo := BuildPanicInfo(rec, 3)
			fmt.Print(panicInfo.Format())
			err = fmt.Errorf("panic: %v", panicInfo.Value)
		}
	}()
	fn()
	return nil
}

// SafeCallWithResult wraps a function call that returns a value with panic recovery
// Returns result and error (error is set if panic occurs)
func SafeCallWithResult[T any](fn func() T) (result T, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			panicInfo := BuildPanicInfo(rec, 3)
			fmt.Print(panicInfo.Format())
			err = fmt.Errorf("panic: %v", panicInfo.Value)
		}
	}()
	result = fn()
	return result, nil
}

// MustRecover should be used as a deferred function to recover from panics
// It will print the panic info and then re-panic
// Use this when you want to log but not suppress the panic
func MustRecover() {
	if rec := recover(); rec != nil {
		panicInfo := BuildPanicInfo(rec, 3)
		fmt.Print(panicInfo.Format())
		panic(panicInfo.Value) // Re-panic
	}
}

// RecoverAndLog recovers from panic and logs it to standard logger
// Use this in deferred functions where you want to suppress panics
func RecoverAndLog() {
	if rec := recover(); rec != nil {
		panicInfo := BuildPanicInfo(rec, 3)
		fmt.Print(panicInfo.Format())
		log.Printf("Recovered from panic: %v", panicInfo.FormatShort())
	}
}

// Catch wraps a function and catches panics, converting them to errors
// Similar to try-catch in other languages
func Catch(fn func()) error {
	return SafeCall(fn)
}

// CatchWithResult wraps a function that returns a value and catches panics
func CatchWithResult[T any](fn func() T) (T, error) {
	return SafeCallWithResult(fn)
}

// Try is an alias for Catch for more intuitive API
func Try(fn func()) error {
	return Catch(fn)
}

// TryWithResult is an alias for CatchWithResult
func TryWithResult[T any](fn func() T) (T, error) {
	return CatchWithResult(fn)
}
