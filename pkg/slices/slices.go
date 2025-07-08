package slices

func AnyCommon[T comparable](sliceA, sliceB []T) bool {
	for _, e := range sliceA {
		if contains(sliceB, e) {
			return true
		}
	}
	return false
}

func contains[T comparable](slice []T, element T) bool {
	for _, subElement := range slice {
		if subElement == element {
			return true
		}
	}
	return false
}
