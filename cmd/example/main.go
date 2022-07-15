package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/theia/pkg/repo"
	"github.com/m1k8/theia/pkg/theia"
)

type config struct {
	IAPI string `json:"IAPI"`
	DAPI string `json:"DAPI"`
}

var token string
var s *discordgo.Session

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"theia-hello": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guild := i.Interaction.GuildID
		channel := i.Interaction.ChannelID

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hello " + channel + ", of server " + guild,
			},
		})
	},
}

var commandDefs = []*discordgo.ApplicationCommand{
	{
		Name:        "theia-hello",
		Description: "Example command.",
	},
}

type DummyRepo struct{}

func (d *DummyRepo) Register(g, c string) error {
	return nil
}

func (d *DummyRepo) Unregister(g string) error {
	return nil
}

func (d *DummyRepo) GetAll() ([]repo.Server, error) {
	return nil, nil
}

func (d *DummyRepo) Get(guildID string) (*repo.Server, error) {
	return nil, nil
}

func main() {

	token = "<insert token here>"

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	log.SetFlags(log.LstdFlags | log.Llongfile)
	if token == "" {
		log.Println("No token provided. Please set it as the DISCORD_API environment variable")
		return
	}

	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile("log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.SetOutput(os.Stdout)
	} else {
		multi := io.MultiWriter(file, os.Stdout)
		log.SetOutput(multi)
	}

	// We need information about servers (which includes their channels),
	// messages and voice states.
	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	// Open the websocket and begin listening.
	err = s.Open()
	if err != nil {
		log.Println("Error opening Discord session: ", err)
		return
	}

	t := theia.NewTheia(s, &DummyRepo{})

	cmdList, cmdDefs := t.Inject("theia-example", commandDefs, commandHandlers)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := cmdDefs[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
	})

	for _, v := range cmdList {
		c, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}

		log.Println("Created " + c.Name)

		defer func() {
			//cleanup
			cmds, _ := s.ApplicationCommands(s.State.User.ID, "")
			for _, cmd := range cmds {
				err = s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
				if err != nil {
					log.Println(fmt.Errorf("error removing %v: %w", cmd.Name, err))
				} else {
					log.Println("Removed " + cmd.Name)
				}
			}
		}()
	}

	log.Println("Ready!")

	defer func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt)
		<-sc
		s.Close()
	}()
}
