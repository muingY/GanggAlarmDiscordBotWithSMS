package DiscordBotCore

import "github.com/bwmarrin/discordgo"

type DiscordBotCore struct {
	botSession *discordgo.Session
}

func (discordBot *DiscordBotCore) Initialize(token string, handler interface{}) (err error) {
	discordBot.botSession, err = discordgo.New("Bot " + token)
	if err != nil {
		return err
	}
	discordBot.botSession.AddHandler(handler)
	discordBot.botSession.Identify.Intents = discordgo.IntentsGuildMessages

	return nil
}
func (discordBot *DiscordBotCore) Destroy() error {
	err := discordBot.botSession.Close()
	return err
}

func (discordBot *DiscordBotCore) Activate() (err error) {
	err = discordBot.botSession.Open()
	if err != nil {
		return err
	}
	return nil
}

func (discordBot *DiscordBotCore) SendChannelMessage(channelId string, msg string) {
	discordBot.botSession.ChannelMessageSend(channelId, msg)
}
