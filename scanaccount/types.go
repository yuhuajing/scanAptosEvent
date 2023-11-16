package scanaccount

const tokenDepositEvent = "0x3::token::DepositEvent"
const tokenWithdrawEvent = "0x3::token::WithdrawEvent"

type NFTInfo struct {
	ID               uint        `gorm:"primary_key"`
	Sequence         int         `gorm:"sequence"`
	Type             string      `json:"type"`
	Amount           int         `json:"amount"`
	Tokenowner       string      `json:"tokenowner"`
	Tokencreator     string      `json:"tokencreator"`
	TokenData        TokenData   `gorm:"embedded"`
	TokenId          TokenDataId // `gorm:"embedded"`
	Relatedhash      string      `json:"relatedhash"`
	Relatedtimestamp uint64      `json:"relatedtimestamp"`
}

type NFTOwner struct {
	ID               uint      `gorm:"primary_key"`
	Amount           int       `json:"amount"`
	Tokenowner       string    `json:"tokenowner"`
	Tokencreator     string    `json:"tokencreator"`
	TokenData        TokenData `gorm:"embedded"`
	Relatedhash      string    `json:"relatedhash"`
	Relatedtimestamp uint64    `json:"relatedtimestamp"`
}

type AccToken struct {
	ID      uint   `gorm:"primary_key"`
	Address string `json:"address"`
	Contype string `json:"contype"`
	Value   string `json:"value"`
}

type AccountAddress struct {
	ID       uint   `gorm:"primary_key"`
	Address  string `json:"address"`
	Sequence int    `json:"sequence"`
}
