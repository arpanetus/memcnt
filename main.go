package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func titleUpdateRoutine(
	bot *tg.BotAPI, 
	cfg tg.ChatConfig, 
	chanPfx string,
	dur time.Duration,
	isDebug bool,
) {
	var prevNum int

	for {
		time.Sleep(dur)
		func() {
			if isDebug {
				log.Printf("[INFO]: starting the title update process")
			}
			
			num, ok := getNum(bot, cfg)
			if !ok {
				return
			}
			if num == prevNum {
				if isDebug {
					log.Printf("[INFO]: num didn't change")
				}
				return
			}

			prevNum = num
			newName := chanPfx + strconv.Itoa(prevNum)
			setName := tg.SetChatTitleConfig{ChatID: cfg.ChatID, Title: newName}

			resp, err := bot.Request(setName)
			if err != nil {
				errPrint(fmt.Errorf("cannot change title: %w", err))
				return
			}
			if string(resp.Result) != "true" {
				errPrint(fmt.Errorf("tried to change the channel title msg but false was the reply"))
				return
			}
			log.Printf("[MSG] changed name to: %s\n", newName)
		}()
	}
}

func handleTitleUpdate(bot *tg.BotAPI, update *tg.Update) {
	if c:=update.ChannelPost; c!=nil && c.NewChatTitle != "" { 
	
		m, err := bot.Request(tg.NewDeleteMessage(update.ChannelPost.Chat.ID, update.ChannelPost.MessageID)) 
		if err!=nil {
			errPrint(fmt.Errorf("cannot delete the new title msg: %w", err))
		} else {
			log.Printf("[INFO]: deleted the older title: %v\n", m)
		}	
	}
}

func removeTitleUpdMsgs(bot *tg.BotAPI, baseUrl string, isWh bool, dur time.Duration) {
	updCfg := tg.NewUpdate(0)
	updCfg.Timeout = int(dur.Seconds())

	if isWh {
		d := tg.DeleteWebhookConfig{DropPendingUpdates: false}

		r, err := bot.Request(d)
		if err!=nil {
			errPrint(fmt.Errorf("cannot delete the webhook: %w", err))
			panic(err)
		}
		log.Printf("[INFO]: successfully deleted the webhook: %s", string(r.Result))

		u, err := url.Parse(baseUrl)
		if err!=nil {
			errPrint(fmt.Errorf("cannot parse the url: %w", err))
			panic(err)
		}
		u.Path = bot.Token

		m, err := bot.Send(tg.WebhookConfig{URL: u})
		if err != nil {
			errPrint(fmt.Errorf("cannot send webhook: %w", err))
			panic(err)
		}
		log.Printf("[INFO]: sent webhook: %s\n", m.Text)

		info, err := bot.GetWebhookInfo()
		if err != nil {
			errPrint(fmt.Errorf("cannot get webhook info: %w", err))
			panic(err)
		}
		if info.LastErrorDate != 0 {
			errPrint(fmt.Errorf("webhook info has last error date: %d", info.LastErrorDate))
			panic(err)
		}

		updates := bot.ListenForWebhook("/"+bot.Token)	
		
		go http.ListenAndServe("0.0.0.0:8080", nil)

		for update := range updates {
			handleTitleUpdate(bot, &update)
		}

	} else {	
		updates := bot.GetUpdatesChan(updCfg)
		
		for update := range updates {
			handleTitleUpdate(bot, &update)
		}
	}

}

var (
	CHANNEL_ID_STR = "CHANNEL_ID"
	CHANNEL_PREFIX = "CHANNEL_PREFIX"
	TELEGRAM_API_TOKEN = "TELEGRAM_API_TOKEN"
	BASE_URL = "BASE_URL"
	IS_WEBHOOKED = "IS_WEBHOOKED"
	CHECK_GETMEMNUM_DUR = "CHECK_GETMEMNUM_DUR"
	IS_DEBUG = "IS_DEBUG"
)

func main() {
	chanId, err := strconv.ParseInt(os.Getenv(CHANNEL_ID_STR), 10, 64)
	if err != nil {
		errPrint(fmt.Errorf("cannot parse the channel_id: %w", err))
		panic(err)
	}

	chanPfx := os.Getenv(CHANNEL_PREFIX)

	token := os.Getenv(TELEGRAM_API_TOKEN)

	bot, err := tg.NewBotAPI(token)
	if err != nil {
		errPrint(fmt.Errorf("cannot initiate bot client: %w", err))
		panic(err)
	}

	baseUrl := os.Getenv(BASE_URL)
	isWh := os.Getenv(IS_WEBHOOKED)!="0"
	isDebug := os.Getenv(IS_DEBUG)!="0"
	
	tmpDur, err := strconv.Atoi(os.Getenv(CHECK_GETMEMNUM_DUR))
	if err!=nil {
		errPrint(fmt.Errorf("cannot parse the check duration: %w", err))
		panic(err)
	}
	dur := time.Millisecond * time.Duration(tmpDur)
	
	bot.Debug = isDebug
	
	chatCfg := tg.ChatConfig{ChatID: chanId}

	chat, err := bot.GetChat(tg.ChatInfoConfig{ChatConfig: chatCfg})
	if err != nil {
		errPrint(fmt.Errorf("cannot get chat: %w", err))
		panic(err)
	}
	log.Printf("[INFO]: init the chat{%s} with title{%s}", chat.UserName, chat.Title)
	
	go titleUpdateRoutine(bot, chatCfg, chanPfx, dur, isDebug)

	removeTitleUpdMsgs(bot, baseUrl, isWh, dur)
}
