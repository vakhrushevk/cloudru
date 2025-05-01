package model

import "time"

type Bucket struct {
	Tokens     int
	Capacity   int
	RefilRate  int
	LastRefill time.Time
}
