package model

type Client struct {
	ID       int    `db:"id"`
	URL      string `db:"url"`
	Rate     int    `db:"rate"`
	Capacity int    `db:"capacity"`
}
