package database

type Account string

func NewAccount(value string) Account {
	return Account(value)
}

type Tx struct {
	From  Account
	To    Account
	Value uint
	Data  string
}

func NewTx(from Account, to Account, value uint, data string) Tx {
	return Tx{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	}
}

func (t Tx) IsReward() bool {
	return t.Data == "reward"
}
