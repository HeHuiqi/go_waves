package account

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var accountsPath string = "accounts.json"

type MainAccount struct {
	Mnenmonic      string `json:"mnenmonic"`
	PrivateKey     string `json:"privateKey"`
	PublicKey      string `json:"publicKey"`
	Address        string `json:"address"`
	TestNetAddress string `json:"testNetAddress"`
}

func (account MainAccount) ToString() string {

	return "Account{\n" +
		"Mnenmonic: " + account.Mnenmonic + "\n" +
		"PrivateKey: " + account.PrivateKey + "\n" +
		"PublicKey: " + account.PublicKey + "\n" +
		"Address: " + account.Address + "\n" +
		"TestNetAddress: " + account.TestNetAddress + "\n" +
		"}\n"
}

type Accounts []MainAccount

var accout *MainAccount
var accouts Accounts

var instanceOnce sync.Once

// 从配置文件中载入json字符串
func LoadAccountsFile(path string) (Accounts, *MainAccount) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panicln("load config conf failed: ", err)
	}
	allAccounts := make(Accounts, 4)
	err = json.Unmarshal(buf, &allAccounts)
	if err != nil {
		log.Panicln("decode config file failed:", string(buf), err)
	}
	mainAccount := &allAccounts[0]

	return allAccounts, mainAccount
}

// 初始化 可以运行多次
func SetAccountsFile(path string) {
	allCccouts, mainAccount := LoadAccountsFile(path)
	accountsPath = path
	accout = mainAccount
	accouts = allCccouts
}

// 初始化，只能运行一次
func InitAccountsFile(path string) *MainAccount {
	if accout != nil && path != accountsPath {
		log.Printf("the config is already initialized, oldPath=%s, path=%s", accountsPath, path)
	}
	instanceOnce.Do(func() {
		allConfigs, MainAccount := LoadAccountsFile(path)
		accountsPath = path
		accout = MainAccount
		accouts = allConfigs
	})

	return accout
}

// 初始化配置文件 为 struct 格式
func Instance() *MainAccount {
	if accout == nil {
		InitAccountsFile(accountsPath)
	}
	return accout
}

// 初始化配置文件 为 []格式
func AllAccounts() Accounts {
	if accout == nil {
		InitAccountsFile(accountsPath)
	}
	return accouts
}

// 获取配置文件路径
func AccountsPath() string {
	return accountsPath
}

func AccountsInit() {
	path := AccountsPath()
	pwd, _ := os.Getwd()
	path = pwd + "/" + path
	println("accounts-path:", path)
	InitAccountsFile(path)

}

// 会自动调用
func init() {
	AccountsInit()
}

func GetMainAccount() MainAccount {
	return GetAccount(0)
}
func GetAccount(index uint8) MainAccount {
	// config.ConfigInit()
	main := AllAccounts()[index]
	return main
}
