// +build !darwin,!linux darwin,!cgo

package watch

import "time"

func startScanner(dir string) {
}

func waitForNextScan(dir string) {
	time.Sleep(700 * time.Millisecond)
}
