package main

import (
	"fmt"
	"strings"

	"github.com/nlopes/slack"
	log "gopkg.in/inconshreveable/log15.v2"
)

func runSlack(rtm *slack.RTM, api *slack.Client) {
	var err error
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			log.Debug("Slack ConnectedEvent", "Infos", ev.Info, "Connection counter", ev.ConnectionCount)

		case *slack.IMCreatedEvent:
			log.Debug("New Slack IMCreatedEvent")
			im := slack.IM{Conversation: slack.Conversation{ID: ev.Channel.ID, IsIM: true, User: ev.User}}
			ims = append(ims, im)

		case *slack.MessageEvent:
			//log.Debug("New Slack MessageEvent", "Event", ev)
			log.Debug("New Slack MessageEvent", "Channel", ev.Channel, "User", ev.User, "Text", ev.Text, "Ts", ev.Timestamp, "Thread TS", ev.ThreadTimestamp)
			text := strings.Replace(ev.Msg.Text, "\n", " ", -1)
			if len(ims) == 0 {
				ims, err = api.GetIMChannels()
				if err != nil {
					log.Warn("In Slack Message", "Couldn't get the IMs", err)
				}
			}
			var isIM bool
			for _, imc := range ims {
				if ev.Channel == imc.Conversation.ID {
					isIM = true
					break
				}
			}
			if !(strings.Contains(text, "!") || strings.Contains(text, "[[")) && !isIM {
				continue
			}
			totalLines.Add(1)
			slackLines.Add(1)
			user, err := api.GetUserInfo(ev.Msg.User)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			var options []slack.RTMsgOption
			if ev.ThreadTimestamp != "" {
				options = append(options, slack.RTMsgOptionTS(ev.ThreadTimestamp))
			}
			toPrint := tokeniseAndDispatchInput(&fryatogParams{slackm: text}, getScryfallCard, getRandomScryfallCard, searchScryfallCard)
			for _, s := range sliceUniqMap(toPrint) {
				if s != "" {
					rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("<@%v>: %v", user.ID, s), ev.Msg.Channel, options...))
				}
			}

		case *slack.PresenceChangeEvent:
			// fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			//fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			log.Error("Slack RTMError", "Error", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Error("Slack InvalidAuthEvent", "event", ev)
			return

		default:
			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}
