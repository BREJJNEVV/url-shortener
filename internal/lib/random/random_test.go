package random_test

import (
	"testing"
	"url-shortener/internal/lib/random"

	"github.com/stretchr/testify/assert"
)

func TestNewRandomString(t *testing.T) {
	testingTable := []struct {
		name     string
		size     int
		expected int
	}{
		{
			name:     "size 1",
			size:     1,
			expected: 1,
		},
		{
			name:     "size 2",
			size:     2,
			expected: 2,
		},
		{
			name:     "size 10",
			size:     10,
			expected: 10,
		},
		{"size 30", 30, 30}, // ОЧ ПРИКОЛЬНАЯ ИНИЦИАЛИЗАЦИЯ СТРУКТУРЫ
		{"size 9999", 999, 999},
	}
	for _, tc := range testingTable { // testCases
		// t.Run Позволяет писать подтесты и запускает каждый тест в своей горутине
		t.Run(tc.name, func(t *testing.T) {
			result := random.NewRandomString(tc.size)
			result2 := random.NewRandomString(tc.size)
			assert.NotEqual(t, result, result2, "Рельультаты совпали!")
			assert.Equal(t, tc.expected, len(result),
				"Для размера %d ожидалась длина %d, получено %d",
				tc.size, tc.expected, len(result),
			)
		})
	}
}
