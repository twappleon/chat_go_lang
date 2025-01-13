// lottery_test.go
package lottery

import (
	"fmt"
	"testing"
)

func TestGenerateWinningNumbers(t *testing.T) {
	numbers := GenerateWinningNumbers(49, 6)
	fmt.Printf("numbers %v\n", numbers)
	if len(numbers) != 6 {
		t.Errorf("Expected 6 numbers, got %d", len(numbers))
	}
}

func TestCheckWinning(t *testing.T) {
	userNumbers := []int{1, 2, 3, 4, 5, 6}
	winningNumbers := []int{1, 2, 3, 4, 5, 6}
	matchCount, matchedNumbers := CheckWinning(userNumbers, winningNumbers)
	fmt.Printf("matchCount %v\n matchCount%v\n", matchCount, matchedNumbers)
	if matchCount != 6 {
		t.Errorf("Expected 6 matches, got %d", matchCount)
	}
	if len(matchedNumbers) != 6 {
		t.Errorf("Expected 6 matched numbers, got %d", len(matchedNumbers))
	}
}

func TestGenerateQuickPick(t *testing.T) {
	numbers := GenerateQuickPick(49, 6)
	if len(numbers) != 6 {
		t.Errorf("Expected 6 numbers, got %d", len(numbers))
	}
}

func TestCheckBigSmall(t *testing.T) {
	numbers := []int{1, 25, 30, 10, 40, 5}
	bigCount, smallCount := CheckBigSmall(numbers)
	if bigCount != 2 {
		t.Errorf("Expected 2 big numbers, got %d", bigCount)
	}
	if smallCount != 4 {
		t.Errorf("Expected 4 small numbers, got %d", smallCount)
	}
}

func TestCheckOddEven(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}
	oddCount, evenCount := CheckOddEven(numbers)
	if oddCount != 3 {
		t.Errorf("Expected 3 odd numbers, got %d", oddCount)
	}
	if evenCount != 3 {
		t.Errorf("Expected 3 even numbers, got %d", evenCount)
	}
}

func TestCheckSumRange(t *testing.T) {
	numbers := []int{10, 20, 30, 40, 50, 60}
	sum, isNormalRange := CheckSumRange(numbers)
	if sum != 210 {
		t.Errorf("Expected sum 210, got %d", sum)
	}
	if isNormalRange {
		t.Errorf("Expected sum to be out of normal range")
	}
}

func TestGetNumberFrequency(t *testing.T) {
	allDraws := [][]int{
		{1, 2, 3, 4, 5, 6},
		{1, 2, 3, 4, 5, 6},
	}
	frequency := GetNumberFrequency(allDraws)
	if frequency[1] != 2 {
		t.Errorf("Expected frequency of 1 to be 2, got %d", frequency[1])
	}
}

func TestCheckConsecutiveNumbers(t *testing.T) {
	numbers := []int{1, 2, 3, 5, 6, 7}
	maxConsecutive := CheckConsecutiveNumbers(numbers)
	if maxConsecutive != 3 {
		t.Errorf("Expected max consecutive to be 3, got %d", maxConsecutive)
	}
}

func TestGetNumberPattern(t *testing.T) {
	numbers := []int{1, 17, 33, 16, 32, 48}
	lowCount, midCount, highCount := GetNumberPattern(numbers)
	if lowCount != 2 {
		t.Errorf("Expected 2 low numbers, got %d", lowCount)
	}
	if midCount != 2 {
		t.Errorf("Expected 2 mid numbers, got %d", midCount)
	}
	if highCount != 2 {
		t.Errorf("Expected 2 high numbers, got %d", highCount)
	}
}
