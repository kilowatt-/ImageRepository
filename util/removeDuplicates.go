package util

// Removes duplicates from an array.
func RemoveDuplicatesFromStringArray(arr []string) []string {
	set := make(map[string]bool)

	for _, k := range arr {
		set[k] = true
	}

	retArr := make([]string, 0, len(set))

	for k, _ := range set {
		retArr = append(retArr, k)
	}

	return retArr
}