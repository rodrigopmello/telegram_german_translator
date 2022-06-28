package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"telegram_german_translator/translation"
)

// Create a struct that mimics the webhook response body
// https://core.telegram.org/bots/api#update
type webhookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		User struct {
			ID int64 `json:"id"`
		}
	} `json:"message"`
}

//Handler This handler is called everytime telegram sends us a webhook event
func Handler(res http.ResponseWriter, req *http.Request) {

	// First, decode the JSON response body
	body := &webhookReqBody{}

	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		fmt.Println("could not decode request body", err)
		return
	}
	log.Println(body)
	// Check if the message contains the command /translate
	// if not, return without doing anything
	switch {
	case strings.Contains(strings.ToLower(body.Message.Text), "/t"):
		text := body.Message.Text[2:]
		if len(text) > 0 {
			if err := handleEvent(body.Message.Chat.ID, body.Message.User.ID, text); err != nil {
				log.Println("error in sending reply:", err)
				return
			}
		}
	default:
		log.Println("Not parsing this text")

	}

	// fmt.Println("reply sent")
}

func handleEvent(chatID, userID int64, text string) error {
	translatedText, err := translateText(chatID, text)
	if err != nil {
		log.Println(err)
	}
	if translation.CheckLanguage(text) {
		err = translatedAudio(chatID, userID, translatedText)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

//The below code deals with the process of sending a response message
// to the user
// Create a struct to conform to the JSON body
// of the send message request
// https://core.telegram.org/bots/api#sendmessage
type translationBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

type audioBody struct {
	Audio []byte `json:"audio"`
}

// translate takes a chatID and sends the translation
func translateText(chatID int64, text string) (string, error) {

	// Create the request body struct
	tr := translationBody{}
	tr.ChatID = chatID

	log.Println("Translating " + text)

	translatedText, err := translation.TranslateText(text)

	if err != nil {
		log.Println(err)
		tr.Text = err.Error()
	}

	log.Println(translatedText)
	tr.Text = translatedText

	// Create the JSON body from the struct
	trBytes, err := json.Marshal(tr)

	if err != nil {
		return "", err
	}
	urlMsg := os.Getenv("TELEGRAM_URL") + "sendMessage" //this url already contains the required token, just to simplify things

	res, err := http.Post(urlMsg, "application/json", bytes.NewBuffer(trBytes))

	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", errors.New("unexpected status" + res.Status)
	}

	return translatedText, nil
}

func translatedAudio(chatID, userID int64, text string) error {
	filename := translation.CreateAudio(userID, text)
	fmt.Println("Sending the Audio")

	chatIDString := strconv.FormatInt(chatID, 10)

	cmd := exec.Command("curl", "-XPOST", os.Getenv("TELEGRAM_URL")+"sendAudio?chat_id="+chatIDString, "-F", "audio=@./"+filename)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("translated phrase: %q\n", out.String())

	return nil
}
