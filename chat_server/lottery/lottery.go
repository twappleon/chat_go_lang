package lottery

import (
	"math/rand"
	"sort"
	"time"
)

// GenerateWinningNumbers generates a slice of winning numbers
// 生成一組中獎號碼
func GenerateWinningNumbers(rangeMax, count int) []int {
	rand.Seed(time.Now().UnixNano())
	numbers := rand.Perm(rangeMax)[:count]
	sort.Ints(numbers)
	return numbers
}

// CheckWinning checks how many numbers match
// 檢查有多少號碼匹配
func CheckWinning(userNumbers, winningNumbers []int) (matchCount int, matchedNumbers []int) {
	matchedNumbers = make([]int, 0)
	for _, num := range userNumbers {
		for _, winNum := range winningNumbers {
			if num == winNum {
				matchCount++
				matchedNumbers = append(matchedNumbers, num)
			}
		}
	}

	// 檢查各種中獎組合
	// Check various winning combinations
	switch matchCount {
	case 6: // 頭獎 - 6個號碼全中
		// First prize - all 6 numbers match
		return matchCount, matchedNumbers
	case 5: // 二獎 - 對中5個號碼
		// Second prize - 5 numbers match
		return matchCount, matchedNumbers
	case 4: // 三獎 - 對中4個號碼
		// Third prize - 4 numbers match
		return matchCount, matchedNumbers
	case 3: // 四獎 - 對中3個號碼
		// Fourth prize - 3 numbers match
		return matchCount, matchedNumbers
	case 2: // 五獎 - 對中2個號碼
		// Fifth prize - 2 numbers match
		return matchCount, matchedNumbers
	default: // 未中獎
		// No prize
		return matchCount, matchedNumbers
	}
}

// GenerateQuickPick generates random numbers for the user
// 生成用戶的隨機號碼
func GenerateQuickPick(rangeMax, count int) []int {
	rand.Seed(time.Now().UnixNano())
	numbers := rand.Perm(rangeMax)[:count]
	sort.Ints(numbers)
	return numbers
}

// CheckBigSmall checks if number is big (>24) or small (<=24)
// 檢查號碼是大於24還是小於等於24
func CheckBigSmall(numbers []int) (bigCount, smallCount int) {
	for _, num := range numbers {
		if num > 24 {
			bigCount++
		} else {
			smallCount++
		}
	}
	return
}

// CheckOddEven checks if number is odd or even
// 檢查號碼是奇數還是偶數
func CheckOddEven(numbers []int) (oddCount, evenCount int) {
	for _, num := range numbers {
		if num%2 == 0 {
			evenCount++
		} else {
			oddCount++
		}
	}
	return
}

// CheckSumRange returns the sum of all numbers and checks if it's in normal range
// 返回所有號碼的總和並檢查是否在正常範圍內
func CheckSumRange(numbers []int) (sum int, isNormalRange bool) {
	for _, num := range numbers {
		sum += num
	}
	// Normal range is typically between 115-185 for 6/49 lottery
	// 對於6/49彩票，正常範圍通常在115-185之間
	isNormalRange = sum >= 115 && sum <= 185
	return
}

// GetNumberFrequency returns a map of how frequently each number appears
// 返回每個號碼出現頻率的映射
func GetNumberFrequency(allDraws [][]int) map[int]int {
	frequency := make(map[int]int)
	for _, draw := range allDraws {
		for _, num := range draw {
			frequency[num]++
		}
	}
	return frequency
}

// CheckConsecutiveNumbers checks for consecutive numbers in selection
// 檢查選擇中的連續號碼
func CheckConsecutiveNumbers(numbers []int) int {
	sort.Ints(numbers)
	maxConsecutive := 1
	currentConsecutive := 1

	for i := 1; i < len(numbers); i++ {
		if numbers[i] == numbers[i-1]+1 {
			currentConsecutive++
			if currentConsecutive > maxConsecutive {
				maxConsecutive = currentConsecutive
			}
		} else {
			currentConsecutive = 1
		}
	}
	return maxConsecutive
}

// GetNumberPattern analyzes number patterns (low/mid/high distribution)
// 分析號碼模式（低/中/高分佈）
func GetNumberPattern(numbers []int) (lowCount, midCount, highCount int) {
	for _, num := range numbers {
		switch {
		case num <= 16:
			lowCount++
		case num <= 32:
			midCount++
		default:
			highCount++
		}
	}
	return
}
