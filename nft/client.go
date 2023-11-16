package nft

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/coming-chat/go-aptos/aptostypes"
	txnBuilder "github.com/coming-chat/go-aptos/transaction_builder"
)

type CollectionData struct {
	// Describes the collection
	Description string `json:"description"`
	// Unique name within this creators account for this collection
	Name string `json:"name"`
	// URL for additional information/media
	Uri string `json:"uri"`
	// Total number of distinct Tokens tracked by the collection
	Count uint64 `json:"count"`
	// Optional maximum number of tokens allowed within this collections
	Maximum uint64 `json:"maximum"`

	Supply uint64 `json:"supply"`
}

type TokenData struct {
	// Unique name within this creators account for this Token's collection
	Collection string `json:"collection"`
	// Describes this Token
	Description string `json:"description"`
	// The name of this Token
	Name string `json:"name"`
	// Optional maximum number of this type of Token.
	Maximum uint64 `json:"maximum"`
	// Total number of this type of Token
	Supply uint64 `json:"supply"`
	/// URL for additional information / media
	Uri string `json:"uri"`
}

type TokenDataId struct {
	/** Token creator address */
	Creator string `json:"creator"`
	/** Unique name within this creator's account for this Token's collection */
	Collection string `json:"collection"`
	/** Name of Token */
	Name string `json:"name"`
}

func (id *TokenDataId) identifier() string {
	return id.Creator + id.Collection + id.Name
}

func (id *TokenDataId) String() string {
	return fmt.Sprintf("{Creator:%v, Collection:%v, Name:%v}", id.Creator, id.Collection, id.Name)
}

type TokenId struct {
	TokenDataId TokenDataId `json:"token_data_id"`
	/** version number of the property map */
	PropertyVersion string `json:"property_version"`
}

type Token struct {
	Id TokenId `json:"id"`
	/** server will return string for u64 */
	Amount string `json:"amount"`
}

type TokenClient struct {
	*aptosclient.RestClient
}

func NewTokenClient(client *aptosclient.RestClient) *TokenClient {
	return &TokenClient{client}
}

/**
 * Queries collection data
 * @param creator Hex-encoded 32 byte Aptos account address which created a collection
 * @param collectionName Collection name
 */
func (c *TokenClient) GetCollectionData(creator txnBuilder.AccountAddress, collectionName string) (*CollectionData, error) {
	collections, err := c.GetAccountResourceByResType(creator.ToShortString(), "0x3::token::Collections", 0)
	if err != nil {
		return nil, err
	}

	handle := ""
	if data, ok := collections.Data["collection_data"].(map[string]interface{}); ok {
		handle, _ = data["handle"].(string)
	}
	body := aptosclient.TableItemRequest{
		KeyType:   "0x1::string::String",
		ValueType: "0x3::token::CollectionData",
		Key:       collectionName,
	}

	out := struct {
		CollectionData
		CountString  string `json:"count"`
		MaxString    string `json:"maximum"`
		SupplyString string `json:"supply"`
	}{}
	err = c.GetTableItem(&out, handle, body, "")
	if err != nil {
		return nil, err
	}
	out.Count, _ = strconv.ParseUint(out.CountString, 10, 64)
	out.Maximum, _ = strconv.ParseUint(out.MaxString, 10, 64)
	out.Supply, _ = strconv.ParseUint(out.SupplyString, 10, 64)
	fmt.Println(out.CollectionData)
	return &out.CollectionData, nil
}

/**
 * Queries token data from collection
 *
 * @param creator Hex-encoded 32 byte Aptos account address which created a token
 * @param collectionName Name of collection, which holds a token
 * @param tokenName Token name
 */
func (c *TokenClient) GetTokenData(creator txnBuilder.AccountAddress, collectionName, tokenName string) (*TokenData, error) {
	collections, err := c.GetAccountResourceByResType(creator.ToShortString(), "0x3::token::Collections", 0)
	if err != nil {
		return nil, err
	}

	handle := ""
	if data, ok := collections.Data["token_data"].(map[string]interface{}); ok {
		handle, _ = data["handle"].(string)
	}
	tokenDataId := TokenDataId{
		Creator:    creator.ToShortString(),
		Collection: collectionName,
		Name:       tokenName,
	}
	body := aptosclient.TableItemRequest{
		KeyType:   "0x3::token::TokenDataId",
		ValueType: "0x3::token::TokenData",
		Key:       tokenDataId,
	}

	out := struct {
		TokenData
		MaxString    string `json:"maximum"`
		SupplyString string `json:"supply"`
	}{}
	err = c.GetTableItem(&out, handle, body, "")
	if err != nil {
		return nil, err
	}
	out.Maximum, _ = strconv.ParseUint(out.MaxString, 10, 64)
	out.Supply, _ = strconv.ParseUint(out.SupplyString, 10, 64)
	out.Collection = collectionName
	//fmt.Println(out.TokenData)
	return &out.TokenData, nil
}

/**
 * Queries token balance for a token account
 * @param account Hex-encoded 32 byte Aptos account address which created a token
 * @param tokenId token id
 */
func (c *TokenClient) GetTokenForAccount(account string, tokenId TokenId) (*Token, error) {
	if tokenId.PropertyVersion == "" {
		tokenId.PropertyVersion = "0"
	}
	tokenStore, err := c.GetAccountResourceByResType(account, "0x3::token::TokenStore", 0)
	if err != nil {
		return nil, err
	}

	handle := ""
	if data, ok := tokenStore.Data["tokens"].(map[string]interface{}); ok {
		handle, _ = data["handle"].(string)
	}
	body := aptosclient.TableItemRequest{
		KeyType:   "0x3::token::TokenId",
		ValueType: "0x3::token::Token",
		Key:       tokenId,
	}

	var out Token
	err = c.GetTableItem(&out, handle, body, "")
	if err != nil {
		if restErr, ok := err.(*aptostypes.RestError); ok && restErr.Code == 404 {
			return &Token{
				Id:     tokenId,
				Amount: "0",
			}, nil
		}
		return nil, err
	}

	return &out, nil
}

type NFTInfo struct {
	TokenData        *TokenData
	TokenId          *TokenDataId
	RelatedHash      string
	RelatedTimestamp uint64
}

func (c *TokenClient) GetAllTokenForAccount(account txnBuilder.AccountAddress) ([]*NFTInfo, error) {
	// 我们需要遍历该用户所有的交易，从中筛选出获得 NFT 的交易，再根据其中的 NFT 信息去查询详细的数据

	owner := account.ToShortString()
	const tokenDepositEvent = "0x3::token::DepositEvent"   // 存入
	const tokenWithdrawEvent = "0x3::token::WithdrawEvent" // 转出

	nftsMap := make(map[string]*NFTInfo, 0) // 这一步执行之后的 info 里面，还未包含 tokenData
	parseNftFromTransaction := func(txn aptostypes.Transaction) (err error) {
		if !txn.Success {
			return
		}
		for _, event := range txn.Events {
			if event.Guid.AccountAddress != owner || (event.Type != tokenDepositEvent && event.Type != tokenWithdrawEvent) {
				continue
			}
			bytes, err := json.Marshal(event.Data)
			if err != nil {
				continue
			}
			token := Token{}
			err = json.Unmarshal(bytes, &token)
			if err != nil {
				continue
			}
			tokenKey := token.Id.TokenDataId.identifier()
			_, exists := nftsMap[tokenKey]
			switch event.Type {
			case tokenWithdrawEvent:
				if exists {
					delete(nftsMap, tokenKey)
				}
			case tokenDepositEvent:
				if exists {
					continue
				}
				nft := &NFTInfo{
					TokenData:        nil,
					TokenId:          &token.Id.TokenDataId,
					RelatedHash:      txn.Hash,
					RelatedTimestamp: txn.Timestamp,
				}
				nftsMap[tokenKey] = nft
			}
		}
		return nil
	}

	const limit = 200
	offset := uint64(0)
	for {
		txns, err := c.GetAccountTransactions(owner, offset, limit)
		if err != nil {
			return nil, err
		}

		for _, txn := range txns {
			err = parseNftFromTransaction(txn)
			if err != nil {
				return nil, err
			}
		}

		if len(txns) < limit {
			break
		} else {
			offset = offset + limit
		}
	}

	nfts := []*NFTInfo{}
	for _, nft := range nftsMap {
		tokenId := nft.TokenId
		creator, err := txnBuilder.NewAccountAddressFromHex(tokenId.Creator)
		if err != nil {
			continue
		}
		tokenData, err := c.GetTokenData(*creator, tokenId.Collection, tokenId.Name)
		if err != nil {
			continue
		}
		nft.TokenData = tokenData
		nfts = append(nfts, nft)
	}

	return nfts, nil
}

type Evolution struct {
	Stage_name_1 string `json:"stage_name_1"`
	Stage_uri_1  string `json:"stage_uri_1"`
	Stage_name_2 string `json:"stage_name_2"`
	Stage_uri_2  string `json:"stage_uri_2"`
	Stage_name_3 string `json:"stage_name_3"`
	Stage_uri_3  string `json:"stage_uri_3"`
	Rarity       string `json:"rarity"`
	Story        string `json:"story"`
}

func (c *TokenClient) GetSpeCollectionData(creator string) (*Evolution, error) {
	collections, err := c.GetAccountResourceByResType(creator, "0x55f710b0b0330e060c41f731ffdd61b846910576bacd0a87be9fd37172012e08::beast_generator::BeastCollection", 0)
	if err != nil {
		return nil, err
	}

	handle := ""
	if data, ok := collections.Data["collections"].(map[string]interface{}); ok {
		handle, _ = data["handle"].(string)
	}
	fmt.Println(handle)
	body := aptosclient.TableItemRequest{
		KeyType:   "u64",
		ValueType: "0x55f710b0b0330e060c41f731ffdd61b846910576bacd0a87be9fd37172012e08::beast_generator::Evolution",
		Key:       strconv.Itoa(154),
	}

	// out := struct {
	// 	Evolution
	// }{}

	out := struct {
		Evolution
		Out_Stage_name_1 string `json:"stage_name_1"`
		Out_Stage_uri_1  string `json:"stage_uri_1"`
		Out_Stage_name_2 string `json:"stage_name_2"`
		Out_Stage_uri_2  string `json:"stage_uri_2"`
		Out_Stage_name_3 string `json:"stage_name_3"`
		Out_Stage_uri_3  string `json:"stage_uri_3"`
	}{}
	err = c.GetTableItem(&out, handle, body, "")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	out.Stage_name_1 = out.Out_Stage_name_1
	out.Stage_name_2 = out.Out_Stage_name_2
	out.Stage_name_3 = out.Out_Stage_name_3
	out.Stage_uri_1 = out.Out_Stage_uri_1
	out.Stage_uri_2 = out.Out_Stage_uri_2
	out.Stage_uri_3 = out.Out_Stage_uri_3
	fmt.Println(out.Evolution)
	return &out.Evolution, nil
}

func (c *TokenClient) GetAllToken(account txnBuilder.AccountAddress) {

	resType, _ := c.GetAccountResources(account.ToShortString(), 0) // 0 means the ;atest version
	for _, res := range resType {
		if StartsWith(res.Type, "0x1::coin::CoinStore") {
			fmt.Println(res.Type)
			coin := res.Data["coin"].(map[string]interface{})
			value := coin["value"].(string)
			balance, _ := big.NewInt(0).SetString(value, 10)
			fmt.Println(balance)
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
