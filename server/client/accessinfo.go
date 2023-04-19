package client

import (
	"time"

	"wsnet2/auth"
)

// AccessInfo : WSNet2への接続に使う情報
type AccessInfo struct {
	LobbyURL  string
	AppId     string
	UserId    string
	MACKey    string
	Bearer    string
	EncMACKey string
}

// GenAccessinfo : AccessInfoを生成
//
// appkeyを知らないクライアントサイドは、この関数を使わずサーバから貰うこと
func GenAccessInfo(lobby, appid, appkey, userid string) (*AccessInfo, error) {
	bearer, err := auth.GenerateAuthData(appkey, userid, time.Now())
	if err != nil {
		return nil, err
	}
	mackey := auth.GenMACKey()
	encmackey, err := auth.EncryptMACKey(appkey, mackey)
	if err != nil {
		return nil, err
	}
	return &AccessInfo{
		LobbyURL:  lobby,
		AppId:     appid,
		UserId:    userid,
		MACKey:    mackey,
		Bearer:    bearer,
		EncMACKey: encmackey,
	}, nil
}
