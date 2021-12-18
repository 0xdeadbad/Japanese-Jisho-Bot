package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/ubermenzchen/LingoWorld-Japanese-Jisho/pkg/jisho"
)

var opts struct {
	Token  string `short:"t" long:"token" description:"Bot token" required:"true"`
	Prefix string `short:"p" long:"prefix" description:"Bot command prefix" required:"true"`
}

// checkErr logs errors if not nil, along with a user-specified trace
func checkErr(err error, trace string, log *logrus.Logger) {
	if err != nil {
		log.WithFields(logrus.Fields{
			"trace": trace,
		}).Error(err)
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	if len(opts.Prefix) > 1 {
		log.Fatalln("Bot command prefix should be only one character")
	}

	log := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	client := disgord.New(disgord.Config{
		ProjectName: "Jisho Bot",
		BotToken:    opts.Token,
		Logger:      log,
		RejectEvents: []string{
			// rarely used, and causes unnecessary spam
			disgord.EvtTypingStart,

			// these require special privilege
			// https://discord.com/developers/docs/topics/gateway#privileged-intents
			disgord.EvtPresenceUpdate,
			disgord.EvtGuildMemberAdd,
			disgord.EvtGuildMemberUpdate,
			disgord.EvtGuildMemberRemove,
		},
		// ! Non-functional due to a current bug, will be fixed.
		Presence: &disgord.UpdateStatusPayload{
			Game: &disgord.Activity{
				Name: fmt.Sprintf("Try %sjisho search house", opts.Prefix),
			},
		},
		// DMIntents: disgord.IntentDirectMessages | disgord.IntentDirectMessageReactions | disgord.IntentDirectMessageTyping,
		// comment out DMIntents if you do not want the bot to handle direct messages

	})

	logFilter, _ := std.NewLogFilter(client)
	filter, _ := std.NewMsgFilter(ctx, client)
	filter.SetPrefix(opts.Prefix)

	// create a handler and bind it to new message events
	// thing about the middlewares are whitelists or passthrough functions.
	client.Gateway().WithMiddleware(
		filter.NotByBot,    // ignore bot messages
		filter.HasPrefix,   // message must have the given prefix
		logFilter.LogMsg,   // log command message
		filter.StripPrefix, // remove the command prefix from the message
	).MessageCreate(func(s disgord.Session, data *disgord.MessageCreate) {
		msg := data.Message

		cmd := strings.Split(msg.Content, " ")

		if cmd[0] != "snake" {
			return
		}

		switch cmd[1] {
		case "search": // whenever the message written is "ping", the bot replies "pong"
			result, err := jisho.JishoSearch(cmd[2])
			if err != nil {
				_, err = msg.Reply(ctx, s, err.Error())
				checkErr(err, "search command", log)
				return
			}
			if len(result.Data) == 0 {
				_, err = msg.Reply(ctx, s, "No information found")
				checkErr(err, "search command", log)
				return
			}

			var entries []string

			for _, entry := range result.Data {
				payload := "```"
				payload += fmt.Sprintf("[ %s ]\n", entry.Slug)
				payload += "[Readings]\n"
				for _, japanese := range entry.Japanese {
					payload += fmt.Sprintf("	╔ Word: %s\n", japanese.Word)
					payload += fmt.Sprintf("	╚ Reading: %s\n", japanese.Reading)
					payload += "\n"
				}
				for _, sense := range entry.Senses {
					types := ""
					tags := ""
					for _, v := range sense.PartsOfSpeech {
						types += fmt.Sprintf("	■ %s\n", v)
					}
					for _, tag := range sense.Tags {
						tags += fmt.Sprintf("	f. %s\n", tag)
					}
					payload += "[English Definitions]\n"
					payload += types
					payload += tags
					for i, definition := range sense.EnglishDefinition {
						if i == 0 {
							payload += fmt.Sprintf("	╔%s\n", definition)
						} else if i < len(sense.EnglishDefinition)-1 {
							payload += fmt.Sprintf("	║%s\n", definition)
						} else {
							payload += fmt.Sprintf("	╚%s\n", definition)
						}
					}
					payload += "\n"

					if len(sense.SeeAlso) > 0 {
						if len(sense.Tags) == 0 {
							payload += "\n"
						}
						payload += "[See Also]\n"
						for _, seeAlso := range sense.SeeAlso {
							payload += fmt.Sprintf("	- %s\n", seeAlso)
						}
						payload += "\n"
					}
				}
				payload += "```"
				entries = append(entries, payload)
			}

			i := 0

			_, err = msg.Reply(ctx, s, fmt.Sprintf("Total of results: %d\n", len(entries)))
			checkErr(err, "search command", log)

			m, err := msg.Reply(ctx, s, entries[i])
			checkErr(err, "search command", log)

			err = m.React(ctx, client, "\U00002B05")
			checkErr(err, "search command", log)
			err = m.React(ctx, client, "\U000027A1")
			checkErr(err, "search command", log)

			client.Gateway().WithMiddleware(
				logFilter.LogMsg,
			).MessageReactionAdd(func(s disgord.Session, h *disgord.MessageReactionAdd) {
				u, err := s.CurrentUser().Get()
				checkErr(err, "search command", log)
				if h.UserID == u.ID {
					return
				}
				if h.MessageID == m.ID {
					if h.PartialEmoji.Name == "\U000027A1" {
						if i == len(entries)-1 {
							i = 0
						} else {
							i++
						}
						_, err := s.Channel(h.ChannelID).Message(h.MessageID).SetContent(entries[i])
						checkErr(err, "search command", log)
					} else if h.PartialEmoji.Name == "\U00002B05" {
						if i == 0 {
							i = len(entries) - 1
						} else {
							i--
						}
						_, err := s.Channel(h.ChannelID).Message(h.MessageID).SetContent(entries[i])
						checkErr(err, "search command", log)
					}
				}
			})

			client.Gateway().WithMiddleware(
				logFilter.LogMsg,
			).MessageReactionRemove(func(s disgord.Session, h *disgord.MessageReactionRemove) {
				u, err := s.CurrentUser().Get()
				checkErr(err, "search command", log)
				if h.UserID == u.ID {
					return
				}
				if h.MessageID == m.ID {
					if h.PartialEmoji.Name == "\U000027A1" {
						if i == len(entries)-1 {
							i = 0
						} else {
							i++
						}
						_, err := s.Channel(h.ChannelID).Message(h.MessageID).SetContent(entries[i])
						checkErr(err, "search command", log)
					} else if h.PartialEmoji.Name == "\U00002B05" {
						if i == 0 {
							i = len(entries) - 1
						} else {
							i--
						}
						_, err := s.Channel(h.ChannelID).Message(h.MessageID).SetContent(entries[i])
						checkErr(err, "search command", log)
					}
				}
			})

		case "shutdown":
			client.Gateway().Disconnect()
			cancel()
		default: // unknown command, bot does nothing.
			return
		}
	})

	// create a handler and bind it to the bot init
	// dummy log print
	client.Gateway().BotReady(func() {
		log.Info("Bot is ready!")
	})

	inviteURL, err := client.BotAuthorizeURL()
	if err != nil {
		panic(err)
	}
	fmt.Println(inviteURL)
	client.Gateway().StayConnectedUntilInterrupted()
	<-ctx.Done()
}
