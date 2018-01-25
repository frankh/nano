package utils

func Reversed(byteArray []byte) (result []byte) {
	for i := len(byteArray) - 1; i >= 0; i-- {
		result = append(result, byteArray[i])
	}
	return result
}
