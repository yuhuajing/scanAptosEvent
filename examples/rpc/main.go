package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/coming-chat/go-aptos/aptostypes"
	"github.com/coming-chat/go-aptos/scanaccount"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const rpcUrl = "https://fullnode.devnet.aptoslabs.com"
const MainnetRestUrl = "https://fullnode.mainnet.aptoslabs.com"

//const rpcUrl = "https://fullnode.mainnet.aptoslabs.com"

func main() {
	ctx := context.Background()
	client, err := aptosclient.Dial(ctx, MainnetRestUrl)
	if err != nil {
		printError(err)
	}

	address := "0x7e5f7bdd454478be1ffe9b66b849efd02359a971aa6a848ceb03bbb5729b3b52"

	dba := scanaccount.Buildconnect()
	dba.AutoMigrate(&scanaccount.NFTInfo{}, &scanaccount.AccToken{}, &scanaccount.NFTOwner{})

	for {
		accountinfo, _ := client.GetAccount(address)
		_sequence := 0
		res := []scanaccount.NFTInfo{}
		dba.Model(&scanaccount.NFTInfo{}).Where("tokenowner = ?", address).Order("sequence desc").Limit(1).Find(&res)
		if len(res) > 0 {
			_sequence = res[0].Sequence + 1
		}
		if int(accountinfo.SequenceNumber) > _sequence {
			scanaccount.GetAllTokenForAccount(dba, client, address, _sequence)
			scanaccount.GetAllToken(dba, client, address)
		} else {
			time.Sleep(5 * time.Minute)
		}
	}
}

func printLine(content string) {
	fmt.Printf("================= %s =================\n", content)
}

func printError(err error) {
	var restError *aptostypes.RestError
	if b := errors.As(err, &restError); b {
		fmt.Printf("code: %d, message: %s, aptos_ledger_version: %d\n", restError.Code, restError.Message, restError.AptosLedgerVersion)
	} else {
		fmt.Printf("err: %s\n", err.Error())
	}
}
