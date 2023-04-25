package wallet

import (
	"encoding/binary"
	"fmt"
	"hqgovaves/account"

	"github.com/mr-tron/base58"

	"github.com/pkg/errors"

	"github.com/tyler-smith/go-bip39"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

const (
	newOpt               = "new"                 // Generate a seed phrase and a wallet
	showOpt              = "show"                // Show existing wallet credentials
	seedPhraseOpt        = "seed-phrase"         // Import a seed phrase
	seedPhraseBase58Opt  = "seed-phrase-base58"  // Import a Base58 encoded seed phrase
	accountSeedBase58Opt = "account-seed-base58" // Import a Base58 encoded account seed

)

const (

	// 助记词的个数 * 11 - 助记词的个数 / 3
	// defaultBitSize = 160
	defaultBitSize = 128
)

type Opts struct {
	seedPhrase        string
	base58SeedPhrase  string
	base58AccountSeed string
}

func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(defaultBitSize)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate random entropy")
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate mnemonic phrase")
	}
	return mnemonic, nil
}

func GenerateOnSeedPhrase(seedPhrase string, n int, scheme byte) (crypto.Digest, crypto.PublicKey, crypto.SecretKey, proto.Address, error) {
	iv := make([]byte, 4)
	binary.BigEndian.PutUint32(iv, uint32(n))
	s := append(iv, seedPhrase...)
	accountSeed, err := crypto.SecureHash(s)
	if err != nil {
		return crypto.Digest{}, crypto.PublicKey{}, crypto.SecretKey{}, nil, errors.Wrap(err, "failed to generate account seed")
	}
	pk, sk, a, err := GenerateOnAccountSeed(accountSeed, scheme)
	if err != nil {
		return crypto.Digest{}, crypto.PublicKey{}, crypto.SecretKey{}, nil, err
	}
	return accountSeed, pk, sk, a, nil
}

func GenerateOnAccountSeed(accountSeed crypto.Digest, scheme proto.Scheme) (crypto.PublicKey, crypto.SecretKey, proto.Address, error) {
	sk, pk, err := crypto.GenerateKeyPair(accountSeed.Bytes())
	if err != nil {
		return crypto.PublicKey{}, crypto.SecretKey{}, nil, errors.Wrap(err, "failed to generate key pair")
	}
	a, err := proto.NewAddressFromPublicKey(scheme, pk)
	if err != nil {
		return crypto.PublicKey{}, crypto.SecretKey{}, nil, errors.Wrap(err, "failed to generate address")
	}
	return pk, sk, a, nil
}

type WalletCredentials struct {
	AccountSeed crypto.Digest
	Pk          crypto.PublicKey
	Sk          crypto.SecretKey
	Address     proto.Address
}

var wrongProgramArguments = errors.New("wrong program arguments were provided")

func GenerateWalletCredentials(
	choice string,
	accountNumber int,
	scheme proto.Scheme,
	opts Opts) (*WalletCredentials, error) {

	var walletCredentials *WalletCredentials

	switch choice {
	case newOpt:
		newSeedPhrase, err := GenerateMnemonic()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate seed phrase")
		}
		accountSeed, pk, sk, address, err := GenerateOnSeedPhrase(newSeedPhrase, accountNumber, scheme)
		if err != nil {
			return nil, err
		}
		walletCredentials = &WalletCredentials{
			AccountSeed: accountSeed,
			Pk:          pk,
			Sk:          sk,
			Address:     address,
		}

		fmt.Printf("Seed Phrase: '%s'\n", newSeedPhrase)
	case seedPhraseOpt:
		if opts.seedPhrase == "" {
			return nil, errors.Wrap(wrongProgramArguments, "no seed phrase was provided")
		}

		accountSeed, pk, sk, address, err := GenerateOnSeedPhrase(opts.seedPhrase, accountNumber, scheme)
		if err != nil {
			return nil, err
		}
		walletCredentials = &WalletCredentials{
			AccountSeed: accountSeed,
			Pk:          pk,
			Sk:          sk,
			Address:     address,
		}
	case seedPhraseBase58Opt:
		if opts.base58SeedPhrase == "" {
			return nil, errors.Wrap(wrongProgramArguments, "no base58 encoded seed phrase was provided")
		}
		b, err := base58.Decode(opts.base58SeedPhrase)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode base58-encoded seed phrase")
		}
		decodedSeedPhrase := string(b)

		accountSeed, pk, sk, address, err := GenerateOnSeedPhrase(decodedSeedPhrase, accountNumber, scheme)
		if err != nil {
			return nil, err
		}
		walletCredentials = &WalletCredentials{
			AccountSeed: accountSeed,
			Pk:          pk,
			Sk:          sk,
			Address:     address,
		}
	case accountSeedBase58Opt:
		if opts.base58AccountSeed == "" {
			return nil, errors.Wrap(wrongProgramArguments, "no base58 account seed was provided")
		}
		accountSeed, err := crypto.NewDigestFromBase58(opts.base58AccountSeed)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode base58-encoded account seed")
		}
		pk, sk, address, err := GenerateOnAccountSeed(accountSeed, scheme)
		if err != nil {
			return nil, err
		}
		walletCredentials = &WalletCredentials{
			AccountSeed: accountSeed,
			Pk:          pk,
			Sk:          sk,
			Address:     address,
		}
	}

	return walletCredentials, nil
}

func CreateWallet(scheme proto.Scheme) (*WalletCredentials, error) {
	var walletCredentials *WalletCredentials
	walletCredentials, err := GenerateWalletCredentials(newOpt, 0, scheme, Opts{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate wallet's credentials")
	}
	if walletCredentials == nil {
		return nil, errors.New("failed to generate wallet's credentials")
	}

	fmt.Printf("Account Seed:   %s\n", walletCredentials.AccountSeed.String())
	fmt.Printf("Public Key:     %s\n", walletCredentials.Pk.String())
	fmt.Printf("Secret Key:     %s\n", walletCredentials.Sk.String())
	fmt.Printf("Address:        %s\n", walletCredentials.Address.String())
	return walletCredentials, nil
}

func CreateWalletMainNet() (*WalletCredentials, error) {
	return CreateWallet(proto.MainNetScheme)
}
func CreateWalletTestNet() (*WalletCredentials, error) {
	return CreateWallet(proto.TestNetScheme)
}

func ImportWallet(mnominc string, scheme proto.Scheme) (*WalletCredentials, error) {
	var walletCredentials *WalletCredentials
	walletCredentials, err := GenerateWalletCredentials(seedPhraseOpt, 0, scheme, Opts{
		seedPhrase: mnominc,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate wallet's credentials")
	}
	if walletCredentials == nil {
		return nil, errors.New("failed to generate wallet's credentials")
	}

	fmt.Printf("Account Seed:   %s\n", walletCredentials.AccountSeed.String())
	fmt.Printf("Public Key:     %s\n", walletCredentials.Pk.String())
	fmt.Printf("Secret Key:     %s\n", walletCredentials.Sk.String())
	fmt.Printf("Address:        %s\n", walletCredentials.Address.String())
	return walletCredentials, nil
}

func ImportWalletMainNet(mnominc string) (*WalletCredentials, error) {
	return ImportWallet(mnominc, proto.MainNetScheme)
}
func ImportWalletTestNet(mnominc string) (*WalletCredentials, error) {
	return ImportWallet(mnominc, proto.TestNetScheme)
}

func ImportWalletFromBs58Privatekey(bs58pri string, scheme proto.Scheme) (crypto.PublicKey, crypto.SecretKey, proto.Address, error) {
	sk, err := crypto.NewSecretKeyFromBase58(bs58pri)
	if err != nil {
		return crypto.PublicKey{}, crypto.SecretKey{}, nil, errors.Wrap(err, "failed to generate key pair")
	}
	pk := crypto.GeneratePublicKey(sk)

	a, err := proto.NewAddressFromPublicKey(scheme, pk)
	if err != nil {
		return crypto.PublicKey{}, crypto.SecretKey{}, nil, errors.Wrap(err, "failed to generate address")
	}
	return pk, sk, a, nil
}
func ImportWalletFromBs58MainNet(bs58pri string) (*WalletCredentials, error) {

	pk, sk, a, err := ImportWalletFromBs58Privatekey(bs58pri, proto.MainNetScheme)
	if err != nil {
		return nil, err
	}
	return &WalletCredentials{
		Pk:      pk,
		Sk:      sk,
		Address: a,
	}, nil

}

func ImportWalletFromBs58TestNet(bs58pri string) (*WalletCredentials, error) {
	pk, sk, a, err := ImportWalletFromBs58Privatekey(bs58pri, proto.MainNetScheme)
	if err != nil {
		return nil, err
	}
	return &WalletCredentials{
		Pk:      pk,
		Sk:      sk,
		Address: a,
	}, nil
}

func CreateWalletTest() {
	CreateWalletMainNet()
	CreateWalletTestNet()
}
func ImportWalletTest() {

	fromAc := account.GetMainAccount()
	seed := fromAc.Mnenmonic
	// seed = "escape era great wreck shop harvest turn antique joy barrel arena congress"
	println(seed)
	ImportWalletMainNet(seed)
	ImportWalletTestNet(seed)
}

func WalletTest() {
	// CreateWalletTest()
	ImportWalletTest()
}
