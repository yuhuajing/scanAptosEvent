package aptosclient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/coming-chat/go-aptos/aptostypes"
)

func Test_GetAccot(t *testing.T) {
	url := MainnetRestUrl
	client := Client(t, url)
	type fields struct {
		chainId int
		c       *http.Client
		rpcUrl  string
		version string
	}
	type args struct {
		address string
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
			args: args{ //https://explorer.aptoslabs.com/account/?network=mainnet
				address: "0x55f710b0b0330e060c41f731ffdd61b846910576bacd0a87be9fd37172012e08",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestClient{
				chainId: tt.fields.chainId,
				c:       tt.fields.c,
				rpcUrl:  tt.fields.rpcUrl,
				version: tt.fields.version,
			}
			accountinfo, err := c.GetAccount(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("RestClient.GetBlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			by, _ := accountinfo.MarshalJSON()
			fmt.Print(string(by))

			// resource, _ := c.GetAccountResourcesInLatestVersion(tt.args.address)
			// for _, v := range resource {
			// 	fmt.Println(v.Type)
			// }
			// resourcev, _ := c.GetAccountResourcesByVersion(tt.args.address, 24577802)
			// for _, v := range resourcev {
			// 	fmt.Println(v.Type)
			// }
			modules, _ := c.GetAccountModules(tt.args.address, 0)
			fmt.Println(len(modules))

		})
	}
}
