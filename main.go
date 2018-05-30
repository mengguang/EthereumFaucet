package main

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"os"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"fmt"
	"flag"
	"github.com/ethereum/go-ethereum/core/types"
	"net/http"
	"log"
)

const defaultWalletPath = "/tmp/ethwallet/"
const defaultWalletPassword = "123qwe"

var walletPath *string
var rpcUrl *string
var globalNonce uint64

func newTestAccount(walletPath string, walletPassword string) {
	wallet := keystore.NewKeyStore(walletPath,
		keystore.LightScryptN, keystore.LightScryptP)

	account, err := wallet.NewAccount(walletPassword)
	if err != nil {
		fmt.Println("Account error:", err)
		os.Exit(1)
	}
	fmt.Println(account.Address.Hex())
}


func sendMoney(toAddress string) error {
	wallet := keystore.NewKeyStore(*walletPath,
		keystore.LightScryptN, keystore.LightScryptP)
	if len(wallet.Accounts()) == 0 {
		return fmt.Errorf("empty wallet, please create account first")
	}
	account := wallet.Accounts()[0]
	wallet.Unlock(account, defaultWalletPassword)

	fmt.Println("account address:", account.Address.Hex())
	client, err := ethclient.Dial(*rpcUrl)
	if err != nil {
		log.Printf("client dial error: %v",err)
		return err
	}
	defer client.Close()
	ctx := context.Background()
	nonce, _ := client.NonceAt(ctx, account.Address, nil)
	if globalNonce < nonce {
		globalNonce = nonce
	}

	fmt.Println("nonce: ", globalNonce)
	var gasLimit uint64 = 21000
	gasPrice := big.NewInt(1)

	amount := big.NewInt(16888)
	amount.Mul(amount,big.NewInt(1000000000000000000))
	tx := types.NewTransaction(globalNonce, common.HexToAddress(toAddress), amount, gasLimit, gasPrice, nil)
	signTx, err := wallet.SignTx(account, tx, nil)
	err = client.SendTransaction(ctx, signTx)
	if err != nil {
		fmt.Println("err:", err)
		return err
	}
	globalNonce++
	return nil
}

func getBalance(address string) *big.Int {
	client, err := ethclient.Dial(*rpcUrl)
	if err != nil {
		log.Printf("client dial error: %v",err)
		return new(big.Int)
	}
	defer client.Close()
	ctx := context.Background()
	fmt.Println(common.HexToAddress(address))
	balance, err := client.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		fmt.Println("Balance error:", err)
		os.Exit(1)
	}
	fmt.Println("Balance: ", balance.String())
	return balance
}

func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	val, ok := r.Form["address"]
	if !ok {
		fmt.Fprintf(w, "Just give me a address!")
		return
	}
	if len(val) != 1 {
		fmt.Fprintf(w, "Just give me ONE address!")
		return
	}
	address := val[0]
	// TODO: check address is valid.
	log.Printf("faucet got address: %v",address)
	amount := getBalance(address)

	fmt.Fprintf(w, "balance: %v",amount.String())
}

func faucetHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	val, ok := r.Form["address"]
	if !ok {
		fmt.Fprintf(w, "Just give me a address!")
		return
	}
	if len(val) != 1 {
		fmt.Fprintf(w, "Just give me ONE address!")
		return
	}
	address := val[0]
	// TODO: check address is valid.
	log.Printf("faucet got address: %v",address)
	err := sendMoney(address)
	if err != nil {
		fmt.Fprintf(w, "something is wrong: %v",err)
		return
	} else {
		fmt.Fprintf(w, "Done! go check your money.") // send data to client side
	}
}

func main() {
	walletPath = flag.String("walletPath", defaultWalletPath, "Wallet storage directory")
	rpcUrl = flag.String("rpcUrl", "http://localhost:8601", "Geth json rpc or ipc url")
	newAccount := flag.Bool("newAccount", false, "Create new account")
	flag.Parse()

	if *newAccount == true {
		newTestAccount(*walletPath, defaultWalletPassword)
		os.Exit(0)
	}
	http.HandleFunc("/faucet", faucetHandler)
	http.HandleFunc("/balance", getBalanceHandler)
	log.Fatal(http.ListenAndServe(":8888", nil))
}
