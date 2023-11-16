package aptosclient

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/coming-chat/go-aptos/aptostypes"
)

func TestRestClient_Getevent(t *testing.T) {
	url := MainnetRestUrl
	client := Client(t, url)

	type args struct {
		address    string
		create_num string
		start      uint64
		limit      uint64
	}
	tests := []struct {
		name      string
		args      args
		wantBlock *aptostypes.Block
		wantErr   bool
	}{
		{
			name: "test with tx",
			args: args{
				address:    "0x55f710b0b0330e060c41f731ffdd61b846910576bacd0a87be9fd37172012e08",
				create_num: strconv.Itoa(2),
				start:      uint64(0),
				limit:      uint64(3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEvents, _ := client.GetEventsByCreationNumber(tt.args.address, tt.args.create_num, tt.args.start, tt.args.limit)
			for _, gotEvent := range gotEvents {
				content, _ := gotEvent.MarshalJSON()
				fmt.Println(string(content))
			}
		})
	}
}

// https://explorer.aptoslabs.com/account/0xa4eb28bacdf267c37e86d1c8352fe68fb782416024f08af803c7269e85d7f95a/resources?network=mainnet
// get the resource first
// get the type
// get the event name
func TestRestClient_GeteventByHandle(t *testing.T) {
	url := MainnetRestUrl
	client := Client(t, url)

	type args struct {
		address string
		//create_num string
		handle string
		field  string
		start  uint64
		limit  uint64
	}
	tests := []struct {
		name      string
		args      args
		wantBlock *aptostypes.Block
		wantErr   bool
	}{
		{
			name: "test with tx",
			args: args{
				address: "0xa4eb28bacdf267c37e86d1c8352fe68fb782416024f08af803c7269e85d7f95a",
				handle:  "0x3::token::TokenStore",
				field:   "withdraw_events",
				//create_num: strconv.Itoa(6),
				start: uint64(1220),
				limit: uint64(5),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEvents, _ := client.GetEventsByEventHandle(tt.args.address, tt.args.handle, tt.args.field, tt.args.start, tt.args.limit)
			fmt.Println(len(gotEvents))
			for _, gotEvent := range gotEvents {
				content, _ := gotEvent.MarshalJSON()
				fmt.Println(string(content))
			}
		})
	}
}
