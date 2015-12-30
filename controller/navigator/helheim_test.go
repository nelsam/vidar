package navigator_test

import (
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commands"
)

type mockPane struct {
	ButtonCalled chan bool
	ButtonOutput struct {
		ret0 chan gxui.Button
	}
	FrameCalled chan bool
	FrameOutput struct {
		ret0 chan gxui.Control
	}
	OnCompleteCalled chan bool
	OnCompleteInput  struct {
		arg0 chan func(commands.Command)
	}
}

func newMockPane() *mockPane {
	m := &mockPane{}
	m.ButtonCalled = make(chan bool, 100)
	m.ButtonOutput.ret0 = make(chan gxui.Button, 100)
	m.FrameCalled = make(chan bool, 100)
	m.FrameOutput.ret0 = make(chan gxui.Control, 100)
	m.OnCompleteCalled = make(chan bool, 100)
	m.OnCompleteInput.arg0 = make(chan func(commands.Command), 100)
	return m
}
func (m *mockPane) Button() gxui.Button {
	m.ButtonCalled <- true
	return <-m.ButtonOutput.ret0
}
func (m *mockPane) Frame() gxui.Control {
	m.FrameCalled <- true
	return <-m.FrameOutput.ret0
}
func (m *mockPane) OnComplete(arg0 func(commands.Command)) {
	m.OnCompleteCalled <- true
	m.OnCompleteInput.arg0 <- arg0
}

type mockCommandExecutor struct {
	ExecuteCalled chan bool
	ExecuteInput  struct {
		arg0 chan commands.Command
	}
}

func newMockCommandExecutor() *mockCommandExecutor {
	m := &mockCommandExecutor{}
	m.ExecuteCalled = make(chan bool, 100)
	m.ExecuteInput.arg0 = make(chan commands.Command, 100)
	return m
}
func (m *mockCommandExecutor) Execute(arg0 commands.Command) {
	m.ExecuteCalled <- true
	m.ExecuteInput.arg0 <- arg0
}

type mockCaller struct {
	CallCalled chan bool
	CallInput  struct {
		arg0 chan func()
	}
	CallOutput struct {
		ret0 chan bool
	}
}

func newMockCaller() *mockCaller {
	m := &mockCaller{}
	m.CallCalled = make(chan bool, 100)
	m.CallInput.arg0 = make(chan func(), 100)
	m.CallOutput.ret0 = make(chan bool, 100)
	return m
}
func (m *mockCaller) Call(arg0 func()) bool {
	m.CallCalled <- true
	m.CallInput.arg0 <- arg0
	return <-m.CallOutput.ret0
}

type mockLocationer struct {
	FileCalled chan bool
	FileOutput struct {
		ret0 chan string
	}
	PosCalled chan bool
	PosOutput struct {
		ret0 chan int
	}
}

func newMockLocationer() *mockLocationer {
	m := &mockLocationer{}
	m.FileCalled = make(chan bool, 100)
	m.FileOutput.ret0 = make(chan string, 100)
	m.PosCalled = make(chan bool, 100)
	m.PosOutput.ret0 = make(chan int, 100)
	return m
}
func (m *mockLocationer) File() string {
	m.FileCalled <- true
	return <-m.FileOutput.ret0
}
func (m *mockLocationer) Pos() int {
	m.PosCalled <- true
	return <-m.PosOutput.ret0
}
