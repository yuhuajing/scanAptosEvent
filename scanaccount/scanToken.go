package scanaccount

import (
	"math/big"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/jinzhu/gorm"
)

func GetAllToken(dba *gorm.DB, c *aptosclient.RestClient, owner string) {
	resType, _ := c.GetAccountResources(owner, 0) // 0 means the ;atest version
	for _, res := range resType {
		if StartsWith(res.Type, "0x1::coin::CoinStore") {
			//fmt.Println(res.Type)
			coin := res.Data["coin"].(map[string]interface{})
			value := coin["value"].(string)
			balance, _ := big.NewInt(0).SetString(value, 10)

			Inserttoken(dba, AccToken{Address: owner, Contype: res.Type, Value: balance.String()})
			//fmt.Println(balance)
		}
	}
}

func StartsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}
