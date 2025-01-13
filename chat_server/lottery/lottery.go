package lottery

import (
	"math/rand"
	"sort"
	"time"
)

// GenerateWinningNumbers generates a slice of winning numbers
func GenerateWinningNumbers(rangeMax, count int) []int {
	rand.Seed(time.Now().UnixNano())
	numbers := rand.Perm(rangeMax)[:count]
	sort.Ints(numbers)
	return numbers
}

// CheckWinning checks how many numbers match
func CheckWinning(userNumbers, winningNumbers []int) int {
	matchCount := 0
	for _, num := range userNumbers {
		for _, winNum := range winningNumbers {
			if num == winNum {
				matchCount++
			}
		}
	}
	return matchCount
}
