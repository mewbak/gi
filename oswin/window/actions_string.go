// Code generated by "stringer -type=Actions"; DO NOT EDIT.

package window

import (
	"fmt"
	"strconv"
)

const _Actions_name = "CloseIconifyResizeMoveFocusDeFocusActionsN"

var _Actions_index = [...]uint8{0, 5, 12, 18, 22, 27, 34, 42}

func (i Actions) String() string {
	if i < 0 || i >= Actions(len(_Actions_index)-1) {
		return "Actions(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Actions_name[_Actions_index[i]:_Actions_index[i+1]]
}

func (i *Actions) FromString(s string) error {
	for j := 0; j < len(_Actions_index)-1; j++ {
		if s == _Actions_name[_Actions_index[j]:_Actions_index[j+1]] {
			*i = Actions(j)
			return nil
		}
	}
	return fmt.Errorf("String %v is not a valid option for type Actions", s)
}
