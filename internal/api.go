package internal

import (
	"bot/entity/site_data"
	"bot/entity/telegram"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RegisterHandlers(router *mux.Router, service Service) {
	router.HandleFunc("/turbine", HandleTelegramWebHook(service)).Methods("POST")
}

func HandleTelegramWebHook(service Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var update telegram.Update
		var message string
		var siteInfo string
		responseSite := site_data.Turbine{}

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			log.Printf("could not decode incoming update %s\n", err.Error())
			return
		}
		if update.UpdateId == 0 {
			log.Printf("invalid update id, got update id = 0")
			err := errors.New("invalid update id of 0 indicates failure to parse incoming update")
			if err != nil {
				return
			}
			return
		}
		log.Printf("update parsed, got %s", &update)
		log.Printf("command arrived, %s", update.Message.Text)

		erro := checkCommand(update.Message.Text)
		if erro == "errore" {
			message = service.PrepareErrorMessageToTelegramChat()
			log.Printf("send to chatId, %s", strconv.Itoa(update.Message.Chat.Id))
			telegramResponseBody, err := service.SendTextToTelegramChat(update.Message.Chat.Id, message)
			if err != nil {
				log.Printf("got error %s from telegram, response body is %s", err.Error(), telegramResponseBody)
			} else {
				log.Printf("turbine infos %s successfully distributed to chat id %d", siteInfo, update.Message.Chat.Id)
			}
			return
		}

		if update.Message.Text == "/start" {
			message = service.PrepareStartMessageToTelegramChat()
			log.Printf("send to chatId, %s", strconv.Itoa(update.Message.Chat.Id))
			telegramResponseBody, err := service.SendTextToTelegramChat(update.Message.Chat.Id, message)
			if err != nil {
				log.Printf("got error %s from telegram, response body is %s", err.Error(), telegramResponseBody)
			} else {
				log.Printf("turbine infos %s successfully distributed to chat id %d", siteInfo, update.Message.Chat.Id)
			}
			return
		} else if update.Message.Text == "/vestas" {
			turbinaVestas, err := service.GetVestasTurbineInfo()
			if err != nil {
				log.Printf("got error: %s from vestas service", err.Error())
				return
			}
			message = service.PrepareVestasTextToTelegramChat(turbinaVestas)
			log.Printf("send to chatId, %s", strconv.Itoa(update.Message.Chat.Id))
			telegramResponseBody, err := service.SendTextToTelegramChat(update.Message.Chat.Id, message)
			if err != nil {
				log.Printf("got error %s from telegram, response body is %s", err.Error(), telegramResponseBody)
			} else {
				log.Printf("turbine infos %s successfully distributed to chat id %d", siteInfo, update.Message.Chat.Id)
			}
			return
		} else {
			a := []string{"1384", "1396", "1397", "1377"}
			b := make([]string, len(a))
			for i, siteId := range a {
				siteInfo, err = service.GetNorthernPowerTurbineInfo(siteId)
				if err != nil {
					log.Printf("got error when calling Northern Power API %s", err.Error())
					return
				}

				err = json.Unmarshal([]byte(siteInfo), &responseSite)
				if err != nil {
					log.Printf("got error %s from parsing response string to response JSON, %s", err.Error(), siteInfo)
				}

				responseSite.DailyPower, _ = service.GetDailyPowerInfo(siteId)

				if responseSite.Site.Device.Faulted == "1" {
					b[i] = service.PrepareScheduledMessageNorthernPowerFaultedToTelegramChat(siteId, responseSite)
				} else if responseSite.Site.Device.Env == "1" {
					b[i] = service.PrepareScheduledMessageNorthernPowerEnvToTelegramChat(siteId, responseSite)
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
					b[i] = service.PrepareScheduledMessageNorthernPowerNoDataToTelegramChat(siteId, responseSite)
					//} else if responseSite.Site.Device.Operating == "0" &&
					//responseSite.Site.Device.Faulted == "0" {
					//b[i] = service.PrepareScheduledMessageNoOperetingToTelegramChat(siteId, responseSite)
				} else {
					b[i] = service.PrepareNorthernPowerTextToTelegramChat(siteId, responseSite)
				}
			}

			if update.Message.Text == "/all" {
				turbinaVestas, err := service.GetVestasTurbineInfo()
				if err != nil {
					log.Printf("got error: %s from vestas service", err.Error())
					return
				}
				message = service.PrepareVestasTextToTelegramChat(turbinaVestas)
				b = append(b, message)
			}
			// Send the punchline back to Telegram
			log.Printf("send to chatId, %s", strconv.Itoa(update.Message.Chat.Id))
			telegramResponseBody, err := service.SendTextToTelegramChat(update.Message.Chat.Id, strings.Join(b, " "))
			if err != nil {
				log.Printf("got error %s from telegram, response body is %s", err.Error(), telegramResponseBody)
			} else {
				log.Printf("turbine infos %s successfully distributed to chat id %d", siteInfo, update.Message.Chat.Id)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func checkCommand(command string) string {
	if command == "/start" || command == "/all" || command == "/vestas" || command == "/northern" {
		return ""
	}

	log.Printf("invalid command, %s", command)
	return "errore"
}
