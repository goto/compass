package generic_helper

// Contains checks if a target item exists in an array of any type.
func Contains[T comparable](arr []T, target T) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}
