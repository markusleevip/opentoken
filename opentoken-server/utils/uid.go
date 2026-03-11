package utils

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var uidRandSeedOnce sync.Once

func GenerateUID() string {
	uidRandSeedOnce.Do(func() {
		rand.Seed(time.Now().UnixNano())
	})

	now := time.Now()
	datePart := now.Format("060102")                  // 年月日
	randomPart := fmt.Sprintf("%06d", rand.Intn(1e6)) // 6位随机数
	return datePart + randomPart
}
