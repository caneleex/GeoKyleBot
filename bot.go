package main

import (
	"context"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"os"
	"os/signal"
	"syscall"
)

var (
	videosChannelId = snowflake.ID(926212070242942986)
)

func main() {
	log.SetLevel(log.LevelInfo)
	log.Info("starting the bot...")
	log.Info("disgo version: ", disgo.Version)

	client, err := disgo.New(os.Getenv("GEOKYLE_TOKEN"),
		bot.WithGatewayConfigOpts(gateway.WithGatewayIntents(discord.GatewayIntentGuildMessages)),
		bot.WithCacheConfigOpts(cache.WithCacheFlags(cache.FlagsNone)),
		bot.WithEventListeners(&events.ListenerAdapter{
			OnComponentInteraction: onButton,
			OnGuildMessageCreate:   onMessage,
		}))
	if err != nil {
		log.Fatal("error while building disgo instance: ", err)
	}

	defer client.Close(context.TODO())

	if client.ConnectGateway(context.TODO()) != nil {
		log.Fatalf("error while connecting to the gateway: %s", err)
	}

	log.Infof("geokyle bot is now running.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-s
}

func onButton(event *events.ComponentInteractionCreate) {
	member := event.Member()
	buttonId := event.Data.CustomID()
	roleId, _ := snowflake.Parse(buttonId.String())
	var hasRole bool
	for _, role := range member.RoleIDs {
		if role == roleId {
			hasRole = true
		}
	}
	rest := event.Client().Rest()
	guildId := *event.GuildID()
	userId := member.User.ID
	var err error
	if hasRole {
		err = rest.RemoveMemberRole(guildId, userId, roleId)
	} else {
		err = rest.AddMemberRole(guildId, userId, roleId)
	}
	messageBuilder := discord.NewMessageCreateBuilder()
	if err != nil {
		event.CreateMessage(messageBuilder.
			SetContentf("❌ There was an error while toggling your role: %s", err).
			SetEphemeral(true).
			Build())
	} else {
		event.CreateMessage(messageBuilder.
			SetContentf("✅ Successfully toggled role <@&%s>.", roleId).
			SetEphemeral(true).
			Build())
	}
}

func onMessage(event *events.GuildMessageCreate) {
	if event.ChannelID != videosChannelId {
		return
	}
	_, err := event.Client().Rest().CrosspostMessage(videosChannelId, event.Message.ID)
	if err != nil {
		log.Errorf("error while crossposting message: %s", err)
	}
}
