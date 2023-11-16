package aptosclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/coming-chat/go-aptos/aptostypes"
)

func Test_chaindata(t *testing.T) {
	ledgerInfo := getChaindata(t)
	content, _ := json.Marshal(ledgerInfo)
	fmt.Println(string(content))
}

func getChaindata(t *testing.T) *aptostypes.LedgerInfo {
	url := MainnetRestUrl
	client := Client(t, url)
	ledgerInfo, err := client.LedgerInfo()
	if err != nil {
		panic(err)
	}

	return ledgerInfo
}

func TestRestClient_GetBlockByHeight(t *testing.T) {
	url := MainnetRestUrl
	client := Client(t, url)
	ledgerInfo := getChaindata(t)

	type fields struct {
		chainId int
		c       *http.Client
		rpcUrl  string
		version string
	}
	type args struct {
		height            string
		with_transactions bool
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantBlock *aptostypes.Block
		wantErr   bool
	}{
		{
			name: "test with tx",
			fields: fields{
				chainId: 1,
				c:       client.c,
				rpcUrl:  url,
				version: VERSION1,
			},
			args: args{
				height:            strconv.Itoa(int(ledgerInfo.BlockHeight)),
				with_transactions: true,
			},
			wantBlock: &aptostypes.Block{BlockHeight: ledgerInfo.BlockHeight},
			wantErr:   false,
		},
		// {
		// 	name: "test without tx",
		// 	fields: fields{
		// 		chainId: 1,
		// 		c:       client.c,
		// 		rpcUrl:  url,
		// 		version: VERSION1,
		// 	},
		// 	args: args{
		// 		height:            "11",
		// 		with_transactions: false,
		// 	},
		// 	wantBlock: &aptostypes.Block{BlockHeight: 11, Transactions: []aptostypes.Transaction{{}}},
		// 	wantErr:   false,
		// },
		// {
		// 	name: "test error",
		// 	fields: fields{
		// 		chainId: 1,
		// 		c:       client.c,
		// 		rpcUrl:  url,
		// 		version: VERSION1,
		// 	},
		// 	args: args{
		// 		height:            "-1",
		// 		with_transactions: true,
		// 	},
		// 	wantBlock: &aptostypes.Block{},
		// 	wantErr:   true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestClient{
				chainId: tt.fields.chainId,
				c:       tt.fields.c,
				rpcUrl:  tt.fields.rpcUrl,
				version: tt.fields.version,
			}
			gotBlock, err := c.GetBlockByHeight(tt.args.height, tt.args.with_transactions)
			if (err != nil) != tt.wantErr {
				t.Errorf("RestClient.GetBlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			content, _ := gotBlock.MarshalJSON()
			fmt.Println(string(content))
			// if gotBlock.BlockHeight != tt.wantBlock.BlockHeight {
			// 	t.Errorf("RestClient.GetBlockByHeight() = %v, want %v", gotBlock, tt.wantBlock)
			// }
			// if len(tt.wantBlock.Transactions) > 0 && len(gotBlock.Transactions) == 0 {
			// 	t.Errorf("RestClient.GetBlockByHeight() transaction len fail 01")
			// }
			// if len(tt.wantBlock.Transactions) == 0 && len(gotBlock.Transactions) != 0 {
			// 	t.Errorf("RestClient.GetBlockByHeight() transaction len fail 02")
			// }
		})
	}
}
