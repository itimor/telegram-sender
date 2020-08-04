package cron

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/runner"

	"github.com/itimor/telegram-sender/config"
	"github.com/itimor/telegram-sender/corp"
	"github.com/itimor/telegram-sender/dataobj"
	"github.com/itimor/telegram-sender/redisc"
)

var (
	semaphore chan int
	tgClient  *corp.Client
)

func SendTelegram() {
	c := config.Get()

	semaphore = make(chan int, c.Consumer.Worker)

	tgClient = corp.New(c.Telegram.Token, c.Telegram.MangoToken)

	for {
		messages := redisc.Pop(1, c.Consumer.Queue)
		if len(messages) == 0 {
			time.Sleep(time.Duration(300) * time.Millisecond)
			continue
		}

		sendTelegrams(messages)
	}
}

func sendTelegrams(messages []*dataobj.Message) {
	for _, message := range messages {
		semaphore <- 1
		go sendTg(message)
	}
}

func sendTg(message *dataobj.Message) {
	defer func() {
		<-semaphore
	}()

	content := genContent(message)

	logger.Info("<-- hashid: %v -->", message.Event.HashId)
	logger.Infof("hashid: %d: endpoint: %s, metric: %s, tags: %s", message.Event.HashId, message.ReadableEndpoint, strings.Join(message.Metrics, ","), message.ReadableTags)

	if count := len(message.Tos); count > 0 {
		for _, tk := range message.Tos {
			err := tgClient.Send(tk, content)
			if err != nil {
				logger.Errorf("send to %s fail:  %v", message.Tos, err)
			}
		}
	}

	logger.Info("<-- /hashid: %v -->", message.Event.HashId)
}

var ET = map[string]string{
	"alert":    "告警",
	"recovery": "恢复",
}

func parseEtime(etime int64) string {
	t := time.Unix(etime, 0)
	return t.Format("2006-01-02 15:04:05")
}

func genContent(message *dataobj.Message) string {
	fp := path.Join(runner.Cwd, "etc", "telegram.tpl")
	t, err := template.ParseFiles(fp)
	if err != nil {
		payload := fmt.Sprintf("InternalServerError: cannot parse %s %v", fp, err)
		logger.Errorf(payload)
		return fmt.Sprintf(payload)
	}

	var body bytes.Buffer
	err = t.Execute(&body, map[string]interface{}{
		"IsAlert":   message.Event.EventType == "alert",
		"Status":    ET[message.Event.EventType],
		"Sname":     message.Event.Sname,
		"Endpoint":  message.ReadableEndpoint,
		"Metric":    strings.Join(message.Metrics, ","),
		"Tags":      message.ReadableTags,
		"Value":     message.Event.Value,
		"Info":      message.Event.Info,
		"Etime":     parseEtime(message.Event.Etime),
		"Elink":     message.EventLink,
		"Slink":     message.StraLink,
		"Clink":     message.ClaimLink,
		"IsUpgrade": message.IsUpgrade,
		"Bindings":  message.Bindings,
		"Priority":  message.Event.Priority,
	})

	if err != nil {
		logger.Errorf("InternalServerError: %v", err)
		return fmt.Sprintf("InternalServerError: %v", err)
	}

	return body.String()
}
