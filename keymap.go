package panelbubble

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type FastKeyMapChecker struct {
	Keys map[string]struct{}
}

func MakeFastKeyMapChecker(keymap interface{}) FastKeyMapChecker {
	checker := FastKeyMapChecker{Keys: make(map[string]struct{})}
	val := reflect.ValueOf(keymap)
	typ := val.Type()
	//fmt.Printf("Type: %v\n", typ)
	for i := 0; i < typ.NumField(); i++ {
		//fmt.Printf("Field %d: %v\n", i, typ.Field(i).Name)
		field := typ.Field(i)
		if field.Type == reflect.TypeOf(key.Binding{}) {
			binding := val.Field(i).Interface().(key.Binding)
			for _, k := range binding.Keys() {
				checker.Keys[k] = struct{}{}
			}
		}
	}
	return checker
}

func (k FastKeyMapChecker) HasKey(msg tea.KeyMsg) bool {
	_, ok := k.Keys[msg.String()]
	return ok
}
