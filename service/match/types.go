package match

type TransactionWithTarget struct {
	Target
}

type Target struct {
	Topic string `json:"topic"`
}
