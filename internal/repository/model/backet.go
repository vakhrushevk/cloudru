// Package model предоставляет структуры для работы с репозиторием
package model

import "time"

// Bucket структура бакета
type Bucket struct {
	Tokens     int
	Capacity   int
	RefilRate  int
	LastRefill time.Time
}
