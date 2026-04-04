package slack

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	stamper "StampCSV/csv"
)

// Bot はSlack Socket Mode Botを表す。
type Bot struct {
	client  *slack.Client
	handler *socketmode.Client
	csvDir  string
}

// NewBot はBotを生成する。
// csvDir: CSVを保存するディレクトリ（動的に変更可能）
func NewBot(csvDir *string) (*Bot, error) {
	appToken := os.Getenv("SLACK_APP_TOKEN")
	botToken := os.Getenv("SLACK_BOT_TOKEN")

	if appToken == "" || botToken == "" {
		return nil, fmt.Errorf("SLACK_APP_TOKEN と SLACK_BOT_TOKEN 環境変数を設定してください")
	}
	if !strings.HasPrefix(appToken, "xapp-") {
		return nil, fmt.Errorf("SLACK_APP_TOKEN は xapp- で始まる必要があります")
	}
	if !strings.HasPrefix(botToken, "xoxb-") {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN は xoxb- で始まる必要があります")
	}

	api := slack.New(botToken,
		slack.OptionAppLevelToken(appToken),
	)
	sm := socketmode.New(api)

	b := &Bot{
		client:  api,
		handler: sm,
	}

	go b.run(sm, csvDir)
	return b, nil
}

func (b *Bot) run(sm *socketmode.Client, csvDir *string) {
	for evt := range sm.Events {
		switch evt.Type {
		case socketmode.EventTypeConnecting:
			log.Println("Slack: 接続中...")
		case socketmode.EventTypeConnected:
			log.Println("Slack: 接続完了")
		case socketmode.EventTypeEventsAPI:
			eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
			if !ok {
				continue
			}
			sm.Ack(*evt.Request)
			b.handleEvent(eventsAPIEvent, csvDir)
		}
	}
}

func (b *Bot) handleEvent(event slackevents.EventsAPIEvent, csvDir *string) {
	if event.Type != slackevents.CallbackEvent {
		return
	}

	innerEvent := event.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		text := strings.ToLower(strings.TrimSpace(ev.Text))
		dir := *csvDir
		if dir == "" {
			b.postMessage(ev.Channel, "CSVディレクトリが設定されていません。UIで選択してください。")
			return
		}

		var stampType string
		if strings.Contains(text, " in") {
			stampType = "in"
		} else if strings.Contains(text, " out") {
			stampType = "out"
		} else {
			b.postMessage(ev.Channel, "使い方: `@StampCSV in` または `@StampCSV out`")
			return
		}

		if err := stamper.Stamp(dir, stampType); err != nil {
			b.postMessage(ev.Channel, "エラー: "+err.Error())
			return
		}

		if stampType == "in" {
			b.postMessage(ev.Channel, "開始を記録しました :white_check_mark:")
		} else {
			b.postMessage(ev.Channel, "終了を記録しました :checkered_flag:")
		}
	}
}

func (b *Bot) postMessage(channel, text string) {
	_, _, err := b.client.PostMessage(channel, slack.MsgOptionText(text, false))
	if err != nil {
		log.Printf("Slack投稿エラー: %v", err)
	}
}

// Start はSocket Modeのイベントループを開始する（ブロッキング）。
func (b *Bot) Start() error {
	return b.handler.Run()
}
