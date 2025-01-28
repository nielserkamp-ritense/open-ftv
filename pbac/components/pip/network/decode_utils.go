package network

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"

func splitKeys(base string) []string {
	var out []string

	var sep rune
	var key []rune

	for _, ch := range base {
		switch ch {
		case '"', '\'':
			if sep == ch {
				out = append(out, string(key))
				key = key[:0]
				sep = 0
			} else if sep != 0 {
				key = append(key, ch)
			} else {
				sep = ch
			}

		case '.':
			if sep != 0 {
				key = append(key, ch)
			} else if len(key) > 0 {
				out = append(out, string(key))
				key = key[:0]
			}

		default:
			key = append(key, ch)
		}
	}

	if len(key) > 0 {
		out = append(out, string(key))
	}
	return out
}

func findElement(keys []string, data any) any {
	if len(keys) == 0 {
		return data
	}

	if m, ok := data.(map[string]any); ok {
		return findElementInMap(keys, m)
	}
	return nil
}

func findElementInMap(keys []string, m map[string]any) any {
	switch len(keys) {
	case 1:
		return m[keys[0]]
	default:
		if m2, ok := m[keys[0]].(map[string]any); ok {
			return findElement(keys[1:], m2)
		}
		return nil
	}
}

func codeOrValue(code string, value any, data map[string]any) any {
	if code != "" {
		if out := data[code]; out != nil {
			return out
		}
	}
	return value
}

func codeOrValueString(code, value string, data map[string]any) string {
	if code != "" {
		if out := convert.AnyToString(data[code]); out != "" {
			return out
		}
	}
	return value
}
