package compare

import "strings"

// StringsEqual return true if all elements in the first list are present in the second list, and vice versa.
//
// The order of the elements can be different as long as all elements exist in both lists.
func StringsEqual(list1, list2 []string) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i := range list1 {
		var ok bool
		for j := range list2 {
			if strings.EqualFold(list1[i], list2[j]) {
				ok = true
				break
			}
		}

		if !ok {
			return false
		}
	}

	return true
}
