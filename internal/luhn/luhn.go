// Package luhn проверяет номера заказов по алгоритму Луна.
package luhn

// Valid проверяет строку из десятичных цифр по алгоритму Луна.
// Пустая строка или нецифровой символ → false.
func Valid(number string) bool {
	if number == "" {
		return false
	}
	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		c := number[i]
		if c < '0' || c > '9' {
			return false
		}
		d := int(c - '0')
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}
