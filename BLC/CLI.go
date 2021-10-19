package BLC

import (
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

// 对 blockchain 的命令行操作进行管理

// CLI client 对象
type CLI struct {
}

// PrintUsage 用法展示
func PrintUsage()  {
	fmt.Println("Usage:")
	// 初始化区块链
	fmt.Printf("\tcreateblockchain -address address -- 创建区块链\n")
	// 添加区块
	fmt.Printf("\taddblock -data DATA-- 添加区块\n")
	// 打印完整的区块信息
	fmt.Printf("\tprintchain -- 输出区块信息\n")
	// 通过命令转账
	fmt.Printf("\tsend -from FROM -to TO -amount AMOUNT -- 发起转账\n")
	// 参数说明
	fmt.Printf("\t转账参数说明\n")
	fmt.Printf("\t\t-from FROM -- 转账源地址\n")
	fmt.Printf("\t\t-to TO -- 转账目标地址\n")
	fmt.Printf("\t\t-amount AMOUNT -- 转账金额\n")
	// 查询余额
	fmt.Printf("\tgetbalance -address FROM -- 查询指定地址的余额\n")
	fmt.Println("\t查询余额参数说明")
	fmt.Printf("\t\t-address -- 查询余额的地址")
}

// getBalance 查询余额
func (cli *CLI) getBalance(from string) {
	// 查找该地址 UTXO
	// 获取区块链对象
	blockchain := BlockchainObject()
	defer blockchain.DB.Close()   // 关闭实例对象
	amount := blockchain.getBalance(from)
	fmt.Printf("\t地址 [%s] 的余额：[%d]\n", from, amount)
}

// send 发起交易
func (cli *CLI) send(from, to, amount []string)  {
	if !dbExit() {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}
	// 获取区块链对象
	blockchain := BlockchainObject()
	defer blockchain.DB.Close()
	blockchain.MineNewBlock(from, to, amount)
}

// createBlockchain 初始化区块链
func (cli *CLI) createBlockchain(address string) {
	CreateBlockChain(address)
}

// addBlock 添加区块
func (cli *CLI) addBlock(txs []*Transaction) {
	if !dbExit() {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}
	blockchain := BlockchainObject()
	blockchain.AddBlock(txs)
}

// printChain 打印完整区块链信息
func (cli *CLI) printChain() {
	if !dbExit() {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}
	blockchain := BlockchainObject()
	blockchain.PrintChain()
}

// dbExit 判断数据库文件是否存在
func dbExit() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		// 数据库文件不存在
		return false
	}
	return true
}

// BlockchainObject 获取一个 blockchain 对象
func BlockchainObject() *BlockChain {
	// 获取 DB
	db, err := bolt.Open(dbName, 0600, nil)
	if nil != err {
		log.Panicf("open the db [%s] failed! %v\n", dbName, err)
	}
	// 获取 Tip
	var tip []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if nil != b {
			tip = b.Get([]byte("1"))
		}
		return nil
	})
	if nil != err {
		log.Panicf("get the blockchain object failed ! %v\n", err)
	}
	return &BlockChain{
		DB: db,
		Tip: tip,
	}
}

// IsValidArgs 参数数量检测函数
func IsValidArgs() {
	if len(os.Args) < 2 {
		PrintUsage()
		// 直接退出
		os.Exit(1)
	}
}

// Run 命令行运行函数
func (cli *CLI) Run() {
	// 检测参数数量
	IsValidArgs()
	// 新建相关命令
	// 添加区块
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	// 输出区块链完整信息
	printchainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	// 创建区块链
	createBLCWithGenesisBlockCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	// 发起交易
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	// 查询余额
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	// 数据参数处理
	// 添加区块
	flagAddBlockArg := addBlockCmd.String("data", "sent 100 btc to player", "添加区块数据")
	// 创建区块时指定的矿工地址
	flagCreateBlockchainArg := createBLCWithGenesisBlockCmd.String("address", "troytan",
		"指定接收系统奖励的矿工地址")
	// 发起交易参数
	flagSendFromArg := sendCmd.String("from", "", "转账源地址")
	flagSendToArg := sendCmd.String("to", "", "转账目标地址")
	flagSendAmountArg := sendCmd.String("amount", "", "转账金额")
	// 查询余额命令行参数
	flagGetBalanceArg := getbalanceCmd.String("address", "", "要查询的地址")
	// 判断命令
	switch os.Args[1] {
	case "getbalance" :
		if err := getbalanceCmd.Parse(os.Args[2:]); nil != err {
			log.Panicf("parse cmd get balance failed %v\n", err)
		}
	case "send":
		if err := sendCmd.Parse(os.Args[2:]); nil != err {
			log.Panicf("parse send failed! %v\n", err)
		}
	case "addblock" :
		if err := addBlockCmd.Parse(os.Args[2:]); nil != err {
			log.Panicf("parse addBlockCmd failed! %v\n", err)
		}
	case "printchain" :
		if err := printchainCmd.Parse(os.Args[2:]); nil != err {
			log.Panicf("parse createBLCWithGenesisBlockCmd failed! %v\n", err)
		}
	case "createblockchain":
		if err := createBLCWithGenesisBlockCmd.Parse(os.Args[2:]); nil != err {
			log.Panicf("parse createBLCWithGenesisBlockCmd failed %v\n", err)
		}
	default:
		// 没有传递任何命令或者传递命令不在上面的命令列表当中
		PrintUsage()
		os.Exit(1)
	}

	// 查询余额
	if getbalanceCmd.Parsed() {
		if "" == *flagGetBalanceArg {
			fmt.Println("请输入查询地址...")
			os.Exit(1)
		}
		cli.getBalance(*flagGetBalanceArg)
	}
	// 发起转账
	if sendCmd.Parsed() {
		if *flagSendFromArg == "" {
			fmt.Println("源地址不能为空...")
			PrintUsage()
			os.Exit(1)
		}
		fmt.Println(*flagSendFromArg)
		if *flagSendToArg == "" {
			fmt.Println("目标地址不能为空...")
			PrintUsage()
			os.Exit(1)
		}
		if *flagSendAmountArg == "" {
			fmt.Println("转账金额不能为空...")
			PrintUsage()
			os.Exit(1)
		}
		fmt.Printf("\tFROM:[%s]\n", JSONToSlice(*flagSendFromArg))
		fmt.Printf("\tTO:[%s]\n", JSONToSlice(*flagSendToArg))
		fmt.Printf("\tAMOUNT:[%s]\n", JSONToSlice(*flagSendAmountArg))
		cli.send(JSONToSlice(*flagSendFromArg), JSONToSlice(*flagSendToArg), JSONToSlice(*flagSendAmountArg))
	}

	// 添加区块命令
	if addBlockCmd.Parsed() {
		if *flagAddBlockArg == "" {
			PrintUsage()
			os.Exit(1)
		}
		cli.addBlock([]*Transaction{})
	}

	// 输出区块链
	if printchainCmd.Parsed() {
		cli.printChain()
	}
	// 创建区块链
	if createBLCWithGenesisBlockCmd.Parsed() {
		if *flagCreateBlockchainArg	== "" {
			PrintUsage()
			os.Exit(1)
		}
		cli.createBlockchain(*flagAddBlockArg)
	}
}