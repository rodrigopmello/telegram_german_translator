package translation

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

//TranslateText functions used to translate pt sentences to german
func TranslateText(text string) (string, error) {
	// text := "The Go Gopher is cute"
	ctx := context.Background()

	dL, err := detectLanguage(text)

	if err != nil {
		log.Println(err)
		return "Language not detected", err
	}
	detectedLang := dL.Language.String()

	targetLanguage := "de"
	if detectedLang == "de" {
		targetLanguage = "pt"
	}
	log.Println("Detected language: ", dL)
	log.Println("Target language: ", targetLanguage)
	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %v", err)
	}

	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	resp, err := client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("Translate: %v", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("Translate returned empty response to text: %s", text)
	}
	return resp[0].Text, nil
}

func detectLanguage(text string) (*translate.Detection, error) {
	ctx := context.Background()
	client, err := translate.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("translate.NewClient: %v", err)
	}
	defer client.Close()
	lang, err := client.DetectLanguage(ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("DetectLanguage: %v", err)
	}
	if len(lang) == 0 || len(lang[0]) == 0 {
		return nil, fmt.Errorf("DetectLanguage return value empty")
	}
	return &lang[0][0], nil
}

// CreateAudio
func CreateAudio(userID int64, text string) string {
	// Instantiates a client.
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	req := texttospeechpb.SynthesizeSpeechRequest{

		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "de",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_OGG_OPUS,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}

	filename := strconv.FormatInt(userID, 10) + "_output.ogg"

	err = ioutil.WriteFile(filename, resp.AudioContent, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Audio content written to file: %v\n", filename)

	return filename
}

//CheckLanguage if its required to send the translated audio
func CheckLanguage(text string) bool {
	dL, err := detectLanguage(text)

	if err != nil {
		log.Println(err)
		return false
	}
	detectedLang := dL.Language.String()

	if detectedLang == "de" {
		return false
	}
	return true
}
