// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import "github.com/nelsam/vidar/commander/bind"

type TabShifter interface {
	ShiftTab(int)
}

type ChangeTab struct {
	shift int
	name  string
}

func NewNextTab() *ChangeTab {
	return &ChangeTab{shift: 1, name: "next-tab"}
}

func NewPrevTab() *ChangeTab {
	return &ChangeTab{shift: -1, name: "prev-tab"}
}

func (t *ChangeTab) Name() string {
	return t.name
}

func (t *ChangeTab) Menu() string {
	return "View"
}

func (t *ChangeTab) Exec(target interface{}) bind.Status {
	shifter, ok := target.(TabShifter)
	if !ok {
		return bind.Waiting
	}
	shifter.ShiftTab(t.shift)
	return bind.Done
}
