package enums

import "strings"

func prepareFromString(in string) string {
	return strings.ToUpper(repl.Replace(strings.TrimSpace(in)))
}

var repl = strings.NewReplacer(" ", "", "-", "", "_", "", "+", "", "&", "", "'", "", "\"", "", "`", "")
