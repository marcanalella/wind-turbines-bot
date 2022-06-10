package internal

import (
	"bot/conf"
	"bot/entity/site_data"
	"encoding/json"
	xj "github.com/basgys/goxml2json"
	"github.com/enescakir/emoji"
	"github.com/go-co-op/gocron"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	GetTurbineInfo(siteId string) (string, error)

	PrepareTextToTelegramChat(command string, turbine site_data.Turbine) string

	PrepareScheduledMessageFaultedToTelegramChat(command string, turbine site_data.Turbine) string

	PrepareScheduledMessageEnvToTelegramChat(siteId string, turbine site_data.Turbine) string

	PrepareScheduledMessageNoDataToTelegramChat(siteId string, turbine site_data.Turbine) string

	PrepareScheduledMessageNoOperetingToTelegramChat(siteId string, turbine site_data.Turbine) string

	PrepareStartMessageToTelegramChat() string

	PrepareErrorMessageToTelegramChat() string

	GetDailyPowerInfo(siteId string) (string, error)

	SendTextToTelegramChat(chatId int, text string) (string, error)

	Schedulednotification()
}

type service struct {
	config conf.Config
}

func NewService(config conf.Config) Service {
	return service{config}
}

func (s service) GetTurbineInfo(siteId string) (string, error) {

	hc := http.Client{}
	form := url.Values{}
	form.Set("username", s.config.NorthernPowerUser)
	form.Set("password", s.config.NorthernPowerPassword)

	req, _ := http.NewRequest("POST", s.config.NorthernPowerApiEnv1+siteId, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	response, err := hc.Do(req)
	if err != nil {
		log.Printf("error while calling Northern Power API %s", err.Error())
		return "", err
	}
	log.Println(response.Status)

	responseJson, err := xj.Convert(response.Body)
	if err != nil {
		log.Printf("could not decode incoming XML response %s", err.Error())
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("failed to retrieve JSON response %s\n", responseJson)
		}
	}(response.Body)

	log.Printf("retrieved JSON response %s", responseJson)
	return responseJson.String(), nil
}

func (s service) GetDailyPowerInfo(siteId string) (string, error) {

	hc := http.Client{}
	form := url.Values{}
	form.Set("username", s.config.NorthernPowerUser)
	form.Set("password", s.config.NorthernPowerPassword)

	from := time.Now().Local()
	fromFormatted := from.Format("2006-01-02")

	req, _ := http.NewRequest("POST", s.config.NorthernPowerApiEnv+siteId+"&from="+fromFormatted+"&to="+fromFormatted, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	response, err := hc.Do(req)
	if err != nil {
		log.Printf("error while calling Northern Power API %s", err.Error())
		return "", err
	}
	log.Println(response.Status)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("failed to retrieve response %v\n", response)
		}
	}(response.Body)

	bodyBytes, err := ioutil.ReadAll(response.Body)
	value := gjson.Get(string(bodyBytes), "energy_kwh")
	log.Printf("retrieved energy_kwh response %s", value)
	return value.String(), nil
}

func (s service) SendTextToTelegramChat(chatId int, text string) (string, error) {

	log.Printf("Sending %s to chat_id: %d", text, chatId)
	response, err := http.PostForm(
		s.config.TelegramApiBaseUrl+s.config.TelegramTokenEnv+s.config.TelegramApiSendMessage,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"via_bot": {"@TurbineNorthernPowerBot"},
			"text":    {text},
		})

	if err != nil {
		log.Printf("error when posting text to the chat: %s", err.Error())
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("error when posting text to the chat: %s", err.Error())
		}
	}(response.Body)

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("error in parsing telegram answer %s", err.Error())
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Printf("Body of Telegram Response: %s", bodyString)

	return bodyString, nil
}

func (s service) PrepareTextToTelegramChat(siteId string, turbine site_data.Turbine) string {

	siteIdReal := ReturnSiteId(siteId)

	return emoji.RadioButton.String() + "TURBINA: " + siteIdReal + "\n\n" +
		emoji.Rocket.String() + "POTENZA TURBINA: " + turbine.Site.Device.Power + "kW" + "\n" +
		emoji.WindFace.String() + "VELOCITA' DEL VENTO: " + turbine.Site.Device.Wndspd + "m/s" + "\n" +
		emoji.Seedling.String() + "ENERGIA GIORNALIERA PRODOTTA: " + turbine.DailyPower + "kWh" + "\n\n" +
		emoji.ExclamationMark.String() + "ULTIMO AGGIORNAMENTO: " + turbine.Site.Ts + "\n\n"
}

func (s service) PrepareScheduledMessageFaultedToTelegramChat(siteId string, turbine site_data.Turbine) string {
	siteIdReal := ReturnSiteId(siteId)
	return emoji.RadioButton.String() + "TURBINA: " + siteIdReal + "\n\n" +
		emoji.Warning.String() + "RIVELATA ANOMALIA: Faulted\n" +
		emoji.Rocket.String() + "POTENZA TURBINA: " + turbine.Site.Device.Power + "kW" + "\n" +
		emoji.WindFace.String() + "VELOCITA' DEL VENTO: " + turbine.Site.Device.Wndspd + "m/s" + "\n" +
		emoji.Seedling.String() + "ENERGIA GIORNALIERA PRODOTTA: " + turbine.DailyPower + "kWh" + "\n\n" +
		emoji.ExclamationMark.String() + "ULTIMO AGGIORNAMENTO: " + turbine.Site.Ts + "\n\n"
}

func (s service) PrepareScheduledMessageEnvToTelegramChat(siteId string, turbine site_data.Turbine) string {
	siteIdReal := ReturnSiteId(siteId)
	return emoji.RadioButton.String() + "TURBINA: " + siteIdReal + "\n\n" +
		emoji.Warning.String() + "RIVELATA ANOMALIA: SPIA BLU " + emoji.BlueCircle.String() + "\n" +
		emoji.Rocket.String() + "POTENZA TURBINA: " + turbine.Site.Device.Power + "kW" + "\n" +
		emoji.WindFace.String() + "VELOCITA' DEL VENTO: " + turbine.Site.Device.Wndspd + "m/s" + "\n" +
		emoji.Seedling.String() + "ENERGIA GIORNALIERA PRODOTTA: " + turbine.DailyPower + "kWh" + "\n\n" +
		emoji.ExclamationMark.String() + "ULTIMO AGGIORNAMENTO: " + turbine.Site.Ts + "\n\n"
}

func (s service) PrepareScheduledMessageNoDataToTelegramChat(siteId string, turbine site_data.Turbine) string {
	siteIdReal := ReturnSiteId(siteId)
	return emoji.RadioButton.String() + "TURBINA: " + siteIdReal + "\n\n" +
		emoji.Warning.String() + "PROBLEMA DI CONNESSIONE: TUTTI I DATI SONO A 0.\n" +
		emoji.Rocket.String() + "POTENZA TURBINA: " + turbine.Site.Device.Power + "kW" + "\n" +
		emoji.WindFace.String() + "VELOCITA' DEL VENTO: " + turbine.Site.Device.Wndspd + "m/s" + "\n" +
		emoji.Seedling.String() + "ENERGIA GIORNALIERA PRODOTTA: " + turbine.DailyPower + "kWh" + "\n\n" +
		emoji.ExclamationMark.String() + "ULTIMO AGGIORNAMENTO: " + time.Now().String() + "\n\n"
}

func (s service) PrepareScheduledMessageNoOperetingToTelegramChat(siteId string, turbine site_data.Turbine) string {
	siteIdReal := ReturnSiteId(siteId)
	return emoji.RadioButton.String() + "TURBINA: " + siteIdReal + "\n\n" +
		emoji.Warning.String() + "TURBINA NON OPERATIVA: ERROR P\n" +
		emoji.Rocket.String() + "POTENZA TURBINA: " + turbine.Site.Device.Power + "kW" + "\n" +
		emoji.WindFace.String() + "VELOCITA' DEL VENTO: " + turbine.Site.Device.Wndspd + "m/s" + "\n" +
		emoji.Seedling.String() + "ENERGIA GIORNALIERA PRODOTTA: " + turbine.DailyPower + "kWh" + "\n\n" +
		emoji.ExclamationMark.String() + "ULTIMO AGGIORNAMENTO: " + time.Now().String() + "\n\n"
}

func (s service) PrepareErrorMessageToTelegramChat() string {
	return emoji.ExclamationMark.String() + "Non ho capito... comando non implementato.\n\n" +
		"Usa il menù del bot per scoprire tutti i comandi disponibili"
}

func (s service) PrepareStartMessageToTelegramChat() string {
	return emoji.WavingHand.String() + " Ciao! io sono @TurbineNorthernPowerBot, e posso aiutarti a monitorare le tue belle turbine!" + "\n\n" +
		emoji.CheckMarkButton.String() + "Le cose da sapere" + "\n\n" +
		emoji.SmallOrangeDiamond.String() + "Il servizio è gratis, ancora in pre-alpha, realizzato da @mariocana."
}

func ReturnSiteId(siteId string) string {
	var siteIdReal string
	switch siteId {
	case "1376":
		siteIdReal = "1042"
	case "1377":
		siteIdReal = "1055"
	case "1384":
		siteIdReal = "1041"
	case "1396":
		siteIdReal = "1046"
	case "1397":
		siteIdReal = "1054"
	}

	return siteIdReal
}

func (s service) Schedulednotification() {
	var message string
	s1 := gocron.NewScheduler(time.UTC)
	_, err := s1.Every(15).Minute().Do(func() {
		a := []string{"1384", "1376", "1396", "1397", "1377"}
		for _, siteId := range a {
			siteInfo, err := s.GetTurbineInfo(siteId)
			if err != nil {
				log.Printf("got error when calling Northern Power API %s", err.Error())
				return
			}

			responseSite := site_data.Turbine{}
			err = json.Unmarshal([]byte(siteInfo), &responseSite)
			if err != nil {
				log.Printf("got error %s from parsing response string to response JSON, %s", err.Error(), siteInfo)
			}
			responseSite.DailyPower, _ = s.GetDailyPowerInfo(siteId)

			if responseSite.Site.Device.Faulted == "1" {
				message = s.PrepareScheduledMessageFaultedToTelegramChat(siteId, responseSite)
			} else if responseSite.Site.Device.Env == "1" {
				message = s.PrepareScheduledMessageEnvToTelegramChat(siteId, responseSite)
			} else if responseSite.Site.Device.Power == "0" &&
				responseSite.Site.Device.Wndspd == "0" &&
				responseSite.Site.Device.Ambtmp == "0" &&
				responseSite.Site.Device.Energy == "0" &&
				responseSite.Site.Device.Yawpos == "0" &&
				responseSite.Site.Device.Yawerr == "0" &&
				responseSite.Site.Device.Operating == "0" &&
				responseSite.Site.Device.Faulted == "0" &&
				responseSite.Site.Device.Warning == "0" &&
				responseSite.Site.Device.Env == "0" &&
				responseSite.Site.Device.Ext == "0" {
				message = s.PrepareScheduledMessageNoDataToTelegramChat(siteId, responseSite)
			} else if responseSite.Site.Device.Operating == "0" &&
				responseSite.Site.Device.Faulted == "0" {
				message = s.PrepareScheduledMessageNoOperetingToTelegramChat(siteId, responseSite)
			}

			b := []int{108781761, 1519990871, 494628308}
			if message != "" {
				for _, chatId := range b {
					// Send the punchline back to Telegram
					log.Printf("send to chatId, %s", strconv.Itoa(chatId))
					telegramResponseBody, err := s.SendTextToTelegramChat(chatId, message)
					if err != nil {
						log.Printf("got error %s from telegram, response body is %s", err.Error(), telegramResponseBody)
					} else {
						log.Printf("turbine infos %s successfully distributed to chat id %d", siteInfo, chatId)
					}
				}
			}
			message = ""
		}
	})
	s1.StartAsync()
	if err != nil {
		log.Printf("error creating job: %v", err)
	}
	_, t := s1.NextRun()
	log.Printf("next run at: %s", t)
}
