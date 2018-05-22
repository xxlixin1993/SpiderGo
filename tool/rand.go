package tool

import (
	"math/rand"
)

// 生成随机整数
func GenerateRangeNum(max int) int {
	randNum := rand.Intn(max)
	return randNum
}