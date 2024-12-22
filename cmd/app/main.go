package main

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"main/internal/voter"
	"main/internal/voterDeleter"
	"main/internal/voterParser"
	"main/pkg/global"
	"main/pkg/types"
	"main/pkg/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func initLog() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}

func processAccounts(
	threads int,
	userAction int,
) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, threads)
	errChan := make(chan error, len(global.AccountsList))

	for _, account := range global.AccountsList {
		wg.Add(1)
		sem <- struct{}{}

		go func(acc types.AccountData) {
			defer wg.Done()
			defer func() { <-sem }()

			var err error
			if userAction == 1 {
				err = voterParser.ParseVotes(acc, util.ProxiesCycler.Next())
			} else if userAction == 2 {
				err = voter.DoVotes(acc, util.ProxiesCycler.Next())
			} else if userAction == 3 {
				err = voterDeleter.DeleteVotes(acc, util.ProxiesCycler.Next())
			}

			if err != nil {
				errChan <- err
			}
		}(account)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			log.Errorf("%v", err)
		}
	}
}

func inputUser(prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return ""
	}
	input = strings.TrimSpace(input)
	return input
}

func handlePanic() {
	if r := recover(); r != nil {
		log.Printf("Unexpected Error: %v", r)
		fmt.Println("Press Enter to Exit..")
		_, err := fmt.Scanln()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(1)
	}
}

func main() {
	defer handlePanic()

	// init log
	var inputData string
	initLog()

	wr, err := os.OpenFile(filepath.Join("log.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Panicf("Error When Opening Log File: %v", err)
	}

	defer func(wr *os.File) {
		err = wr.Close()
		if err != nil {
			log.Panicf("Error When Closing Log File: %v", err)
		}
	}(wr)
	mw := io.MultiWriter(os.Stdout, wr)
	log.SetOutput(mw)

	// init proxies
	err = util.InitProxies(filepath.Join("config", "proxies.txt"))
	if err != nil {
		log.Panicf("Error initializing proxies: %v", err)
	}
	// --> init
	err = util.ReadJsonFile(filepath.Join("config", "const.json"),
		&global.Const)

	if err != nil {
		log.Panicf("Error reading const.json: %v", err)
	}

	accountsListString, err := util.ReadFileByRows(filepath.Join("config", "accounts.txt"))

	if err != nil {
		log.Panicf("Error Reading Accounts List File: %v", err.Error())
	}

	global.AccountsList, err = util.GetAccounts(accountsListString)

	if err != nil {
		log.Panicf(err.Error())
	}

	fmt.Printf("Successfully Loaded %d Accounts / %d Proxies", len(global.AccountsList), len(util.Proxies))

	inputData = inputUser("\n\n1. Parse Accounts Votes\n2. Projects Voter\n3. Votes Deleter\nEnter Your Action: ")

	userAction, err := strconv.Atoi(inputData)

	if err != nil {
		log.Panicf("Wrong User Action Number: %s", inputData)
	}

	inputData = inputUser("Threads: ")
	threads, err := strconv.Atoi(inputData)

	if err != nil {
		log.Panicf("Wrong Threads Number: %s", inputData)
	}

	fmt.Println()

	processAccounts(threads, userAction)

	log.Printf("The Work Has Been Successfully Finished")
	inputUser("\nPress Enter to Exit..")
}
