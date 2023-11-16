package scanaccount

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/coming-chat/go-aptos/aptostypes"
	txnBuilder "github.com/coming-chat/go-aptos/transaction_builder"
	"github.com/jinzhu/gorm"
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
	//fmt.Println(out.CollectionData)
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

func GetAllTokenForAccount(dba *gorm.DB, c *aptosclient.RestClient, owner string, start int) error {
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
			_amount, _ := strconv.Atoi(token.Amount)
			nft := &NFTInfo{
				Sequence:         int(txn.SequenceNumber),
				Type:             event.Type,
				Amount:           _amount,
				Tokenowner:       owner,
				TokenData:        TokenData{},
				TokenId:          token.Id.TokenDataId,
				Relatedhash:      txn.Hash,
				Relatedtimestamp: txn.Timestamp,
			}

			tokenId := nft.TokenId
			creator, err := txnBuilder.NewAccountAddressFromHex(tokenId.Creator)
			if err != nil {
				continue
			}
			tokenData, err := NewTokenClient(c).GetTokenData(*creator, tokenId.Collection, tokenId.Name)
			if err != nil {
				continue
			}
			nft.TokenData = *tokenData

			Insertnft(dba, nft)
		}
		return nil
	}

	const limit = 100
	offset := uint64(start)
	for {
		txns, err := c.GetAccountTransactions(owner, offset, limit)
		if err != nil {
			return err
		}

		for _, txn := range txns {
			err = parseNftFromTransaction(txn)
			if err != nil {
				return err
			}
		}

		if len(txns) < limit {
			break
		} else {
			offset = offset + limit
		}
	}

	return nil
}
