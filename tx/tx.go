package tx

import (
	"hqgovaves/account"

	"encoding/json"
	"strings"

	"github.com/mr-tron/base58/base58"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

func addressFromString(s string) (proto.WavesAddress, error) {
	ab, err := base58.Decode(s)
	if err != nil {
		return proto.WavesAddress{}, err
	}
	a := proto.WavesAddress{}
	copy(a[:], ab)
	return a, nil
}
func recipientFromString(s string) (proto.Recipient, error) {
	if strings.HasPrefix(s, proto.AliasPrefix) {
		a, err := proto.NewAliasFromString(s)
		if err != nil {
			return proto.Recipient{}, err
		}
		return proto.NewRecipientFromAlias(*a), nil
	}
	addr, err := addressFromString(s)
	if err != nil {
		return proto.Recipient{}, err
	}
	return proto.NewRecipientFromAddress(addr), nil
}

func TxTest() {
	TestTransferWithSig()
}

func TestTransferWithSig() {

	fromAc := account.GetMainAccount()
	toAc := account.GetMainAccount()

	priKeybs58 := fromAc.PrivateKey
	pubKeybs58 := fromAc.PublicKey

	//注意网络配置
	scheme := proto.TestNetScheme
	toInfo := struct {
		recipient string // 接收地址
		amount    uint64 // 发送数量
		fee       uint64 // 交易fee
		att       string // 备注
	}{toAc.TestNetAddress, 1000000, 10, "The WAVES Transfer"}

	spk, _ := crypto.NewPublicKeyFromBase58(pubKeybs58)
	rcp, err := recipientFromString(toInfo.recipient)
	if err != nil {
		panic(err)
	}
	a, err := proto.NewOptionalAssetFromString("WAVES")
	if err != nil {
		panic(err)
	}
	att := []byte(toInfo.att)
	tx := proto.NewUnsignedTransferWithSig(spk, *a, *a, 0, toInfo.amount, toInfo.fee, rcp, att)
	_, err = tx.Validate(scheme)
	if err != nil {
		panic(err)
	}
	sk, _ := crypto.NewSecretKeyFromBase58(priKeybs58)
	tx.Sign(scheme, sk)

	bts, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}
	println("bts", string(bts))

	/*
		//官方不提供apikey,
		rpcUrl := "https://testnodes.wavesnodes.com"
		apiKey := ""
		txs := client.NewTransactions(client.Options{
			BaseUrl: rpcUrl,
			Client:  &http.Client{Timeout: 3 * time.Second},
			ApiKey:  apiKey,
		})
		rsp, err := txs.Broadcast(context.Background(), tx)
		if err != nil {
			panic(err)
		}
		println("rsp", rsp)
	*/

}
