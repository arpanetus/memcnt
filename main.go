package main

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
	"time"
)

func errPrint(s error) {
	log.Printf("[ERROR] %v\n", s)
}

func getNum(bot *tg.BotAPI, cfg tg.ChatConfig) (int, bool) {
	num, err := bot.GetChatMembersCount(tg.ChatMemberCountConfig{ChatConfig: cfg})
	if err != nil {
		errPrint(fmt.Errorf("cannot get number of users: %w", err))
		return 0, false
	}
	return num, true
}

func main() {
	chanId, err := strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)
	if err != nil {
		errPrint(fmt.Errorf("cannot parse the channel_id: %w", err))
		panic(err)
	}

	chanPfx := os.Getenv("CHANNEL_PREFIX")

	bot, err := tg.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		errPrint(fmt.Errorf("cannot initiate bot client: %w", err))
		panic(err)
	}

	bot.Debug = true

	chatCfg := tg.ChatConfig{ChatID: chanId}

	chat, err := bot.GetChat(tg.ChatInfoConfig{ChatConfig: chatCfg})
	if err != nil {
		errPrint(fmt.Errorf("cannot get chat: %w", err))
		panic(err)
	}
	log.Printf("[INFO]: init the chat{%s} with title{%s}", chat.UserName, chat.Title)

	prevNum, _ := getNum(bot, chatCfg)

	for {
		time.Sleep(time.Second * 30)
		func() {
			log.Printf("[INFO]: starting the title update process")
			num, ok := getNum(bot, chatCfg)
			if !ok {
				return
			}
			if num == prevNum {
				log.Printf("[INFO]: num didn't change")
				return
			}

			prevNum = num
			newName := chanPfx + strconv.Itoa(prevNum)
			setName := tg.SetChatTitleConfig{ChatID: chatCfg.ChatID, Title: newName}

			resp, err := bot.Request(setName)
			if err != nil {
				errPrint(fmt.Errorf("cannot change title: %w", err))
				return
			}
			if string(resp.Result) != "true" {
				errPrint(fmt.Errorf("tried to delete the prev msg but false was the reply", err))
				return
			}
			log.Printf("[MSG] changed name to: %s\n", newName)
		}()
	}

}
