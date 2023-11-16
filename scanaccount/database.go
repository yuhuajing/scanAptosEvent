package scanaccount

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type MysqlConFig struct {
	Addr            string
	Port            int
	Db              string
	Username        string
	Password        string
	MaxIdealConn    int
	MaxOpenConn     int
	ConnMaxLifetime int
}

var MysqlCon = MysqlConFig{
	"127.0.0.1",
	3306,
	"aptosaccCoin",
	"root",
	"123456",
	10,
	256,
	600,
}

func Buildconnect() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&timeout=%s",
		MysqlCon.Username, MysqlCon.Password, MysqlCon.Addr, MysqlCon.Port, MysqlCon.Db, "10s")
	//mysql connection
	dba, err := gorm.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Connect error:%s\n", err)
	}
	return dba
}

func Insertnft(dba *gorm.DB, nftdata *NFTInfo) {
	res := dba.Model(&NFTInfo{}).Where("type = ? AND tokenowner = ? AND  tokencreator = ? AND collection = ? AND name = ? AND relatedhash=? AND relatedtimestamp=?", nftdata.Type, nftdata.Tokenowner, nftdata.TokenId.Creator, nftdata.TokenId.Collection, nftdata.TokenId.Name, nftdata.Relatedhash, nftdata.Relatedtimestamp).First(&NFTInfo{})
	if res.RowsAffected == 0 {
		dba.Create(&NFTInfo{
			Sequence:         nftdata.Sequence,
			Type:             nftdata.Type,
			Amount:           nftdata.Amount,
			Tokenowner:       nftdata.Tokenowner,
			Tokencreator:     nftdata.TokenId.Creator,
			TokenData:        nftdata.TokenData,
			Relatedhash:      nftdata.Relatedhash,
			Relatedtimestamp: nftdata.Relatedtimestamp,
		})
		InsertnftOwner(dba, nftdata)
	}
}

func InsertnftOwner(dba *gorm.DB, nftdata *NFTInfo) {
	res := NFTOwner{}
	dba.Model(&NFTOwner{}).Where("tokenowner = ? AND  tokencreator = ? AND collection = ? AND name = ?", nftdata.Tokenowner, nftdata.TokenId.Creator, nftdata.TokenId.Collection, nftdata.TokenId.Name).Find(&res)

	switch nftdata.Type {
	case tokenWithdrawEvent:
		if res.Amount-nftdata.Amount == 0 {
			dba.Model(&NFTOwner{}).Where("tokenowner = ? AND  tokencreator = ? AND collection = ? AND name = ?", nftdata.Tokenowner, nftdata.TokenId.Creator, nftdata.TokenId.Collection, nftdata.TokenId.Name).Delete(&NFTOwner{})
		} else {
			dba.Model(&NFTOwner{}).Where("tokenowner = ? AND  tokencreator = ? AND collection = ? AND name = ?", nftdata.Tokenowner, nftdata.TokenId.Creator, nftdata.TokenId.Collection, nftdata.TokenId.Name).Update("amount", res.Amount-nftdata.Amount)
		}

	case tokenDepositEvent:
		if res.Amount == 0 {
			dba.Create(&NFTOwner{
				Amount:           nftdata.Amount,
				Tokenowner:       nftdata.Tokenowner,
				Tokencreator:     nftdata.TokenId.Creator,
				TokenData:        nftdata.TokenData,
				Relatedhash:      nftdata.Relatedhash,
				Relatedtimestamp: nftdata.Relatedtimestamp,
			})
		} else {
			dba.Model(&NFTOwner{}).Where("tokenowner = ? AND  tokencreator = ? AND collection = ? AND name = ?", nftdata.Tokenowner, nftdata.TokenId.Creator, nftdata.TokenId.Collection, nftdata.TokenId.Name).Update("amount", res.Amount+nftdata.Amount)
		}
	}
}

func Inserttoken(dba *gorm.DB, coindata AccToken) {
	res := dba.Model(&AccToken{}).Where("address = ? AND contype = ?", coindata.Address, coindata.Contype).First(&AccToken{})
	if res.RowsAffected == 0 {
		dba.Create(&AccToken{
			Address: coindata.Address,
			Contype: coindata.Contype,
			Value:   coindata.Value,
		})

	} else {
		dba.Model(&AccToken{}).Where("address = ? AND contype = ?", coindata.Address, coindata.Contype).Update("value", coindata.Value)
	}
}
