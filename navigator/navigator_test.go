// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator_test

import (
	"testing"
	"time"

	"github.com/a8m/expect"
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/drivers/gl"
	"github.com/nelsam/gxui/themes/dark"
	"github.com/nelsam/vidar/navigator"
)

func TestNavigator(t *testing.T) {
	expect := expect.New(t)

	mockCommandExecutor := newMockCommandExecutor()

	driver := startDriver()
	defer driver.Terminate()

	theme := dark.CreateTheme(driver)

	var nav *navigator.Navigator
	driver.CallSync(func() {
		nav = navigator.New(driver, theme, mockCommandExecutor)
	})

	expect(nav.Elements()).To.Have.Len(0)
	expect(nav.Buttons().Children()).To.Have.Len(0)
	expect(nav.Children()).To.Have.Len(1)

	mockPane := newMockPane()

	var button gxui.Button
	driver.CallSync(func() {
		button = theme.CreateButton()
	})
	mockPane.ButtonOutput.ret0 <- button

	nav.Add(mockPane)

	waitFor(func() interface{} { return len(nav.Elements()) }, 1, time.Second)
	expect(nav.Elements()).To.Have.Len(1)
	expect(nav.Buttons().Children()).To.Have.Len(1)
	expect(nav.Children()).To.Have.Len(1)

	var frame gxui.Control
	driver.CallSync(func() {
		frame = theme.CreateLinearLayout()
	})
	mockPane.FrameOutput.ret0 <- frame

	driver.CallSync(func() {
		button.Click(gxui.MouseEvent{})
	})
	clicked := waitFor(func() interface{} { return len(mockPane.FrameCalled) }, 1, time.Second)
	expect(clicked).To.Equal(false)
	expect(mockPane.FrameCalled).To.Have.Len(0)

	driver.CallSync(func() {
		button.Click(gxui.MouseEvent{Button: gxui.MouseButtonLeft})
	})
	waitFor(func() interface{} { return len(mockPane.FrameCalled) }, 1, time.Second)
	expect(mockPane.FrameCalled).To.Have.Len(1)
	expect(nav.Children()).To.Have.Len(2)

	mockPane.FrameOutput.ret0 <- frame
	driver.CallSync(func() {
		button.Click(gxui.MouseEvent{Button: gxui.MouseButtonLeft})
	})
	waitFor(func() interface{} { return len(mockPane.FrameCalled) }, 2, time.Second)
	expect(mockPane.FrameCalled).To.Have.Len(2)
	expect(nav.Children()).To.Have.Len(1)
}

// startDriver is just sugar to get a UI goroutine running and return
// the corresponding driver.
func startDriver() gxui.Driver {
	// gl.StartDriver seems to run on the thread that it is given, so
	// we need to kick it off in a separate goroutine.
	drivers := make(chan gxui.Driver)
	go gl.StartDriver(func(driver gxui.Driver) {
		drivers <- driver
	})
	return <-drivers
}

// waitFor is a temporary workaround while I work on implementing
// functionality akin to gomega's Eventually/Consistently in
// a8m/expect.
func waitFor(caller func() interface{}, expected interface{}, timeoutAfter time.Duration) bool {
	timeout := time.After(timeoutAfter)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if caller() == expected {
				return true
			}
		case <-timeout:
			return false
		}
	}
}
