package transformation

import (
	"fmt"
	"sort"
	"strings"
)

type (
	Errors map[string]error
)

func (es Errors) Error() string {
	if len(es) == 0 {
		return ""
	}

	keys := make([]string, len(es))
	i := 0
	for key := range es {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	var s strings.Builder
	for i, key := range keys {
		if i > 0 {
			s.WriteString("; ")
		}
		if errs, ok := es[key].(Errors); ok {
			fmt.Fprintf(&s, "%v: (%v)", key, errs)
			continue
		}

		fmt.Fprintf(&s, "%v: %v", key, es[key].Error())
	}
	s.WriteString(".")

	return s.String()
}
