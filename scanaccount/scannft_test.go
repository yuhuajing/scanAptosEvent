package scanaccount

import (
	"context"
	"fmt"
	"testing"

	"github.com/coming-chat/go-aptos/aptosclient"
	txnBuilder "github.com/coming-chat/go-aptos/transaction_builder"
)

const (
	MainnetRestUrl = "https://fullnode.mainnet.aptoslabs.com"
	TestnetRestUrl = "https://testnet.aptoslabs.com"
	DevnetRestUrl  = "https://fullnode.devnet.aptoslabs.com"
)

const (
	nftCreator = "0x305a97874974fdb9a7ba59dc7cab7714c8e8e00004ac887b6e348496e1981838"
	nftOwner   = "0x37ccac35f1d5a11773d14e47786f112941eb92aaea3295a3487e1b6dc3810b2a"

	nftCollectionName = "Aptos Names V1"
	nftTokenNameOwned = "0-0.apt"
)

var (
	restClient, _ = aptosclient.Dial(context.Background(), MainnetRestUrl)
	tokenClient   = NewTokenClient(restClient)
)

func TestGetTokenForAccount(t *testing.T) {
	tokenId := TokenId{
		TokenDataId: TokenDataId{ //0xd2c3b86a43bd2d3dcb3c27f9360f714124405f93af5fdb94210f42aff3ce79ff
			Creator:    "0xd2c3b86a43bd2d3dcb3c27f9360f714124405f93af5fdb94210f42aff3ce79ff",
			Collection: "W&W Beast",
			Name:       "Blazeplume #10",
		},
	}
	owner, _ := txnBuilder.NewAccountAddressFromHex(tokenId.TokenDataId.Creator)
	tr, _ := tokenClient.GetTokenData(*owner, tokenId.TokenDataId.Collection, tokenId.TokenDataId.Name)
	fmt.Println(tr)

	//t.Log(data.Id.TokenDataId)
}

// func TestGetAllToken(t *testing.T) {
// 	owner, _ := txnBuilder.NewAccountAddressFromHex("0x7e5f7bdd454478be1ffe9b66b849efd02359a971aa6a848ceb03bbb5729b3b52")
// 	tokenClient.GetAllToken(*owner)
// }
