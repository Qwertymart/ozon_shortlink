package base63

import "strings"

const (
	alphabet  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	base      = uint64(len(alphabet))
	resultLen = 10
)

func Encode(n uint64) (string, error) {
	var sb strings.Builder

	// генерация символов
	for n > 0 {
		if err := sb.WriteByte(alphabet[n%base]); err != nil{
			return "", err
		}
		n /= base
	}

	result := sb.String()

	// дополнение строки zero-value значением(в нашем случа это "a") до len = 10
	for len(result) < 10 {
		result += string(alphabet[0])
	}

	return result, nil
}
