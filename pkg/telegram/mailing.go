package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Rosya-edwica/api.edwica/config"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
)

const programName = "api.edwica (golang)"

var cfg *config.Telegram

func init() {
	var err error
	cfg, err = config.LoadTelegramConfig()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
}

func ErrorMessageMailing(text string) {
	Mailing(programName + "\nОШИБКА🛑\n" + text)
}

func SuccessMessageMailing(text string) {
	Mailing(programName + "\nУСПЕШНО✅\n" + text)
}

func Mailing(text string) {
	for _, chat := range cfg.Chats {
		SendMessage(text, chat)
	}
}

func SendMessage(text string, chatId string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.Token)
	params := map[string]string{
		"chat_id": chatId,
		"text":    text,
	}
	body, _ := json.Marshal(params)
	response, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		logger.Log.Info("Не удалось отправить сообщение чату:" + chatId)
	}
	response.Body.Close()
}
