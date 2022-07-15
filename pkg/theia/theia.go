package theia

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/theia/pkg/repo"
)

type Theia struct {
	repo repo.Repo
	s    *discordgo.Session
}

func NewTheia(session *discordgo.Session, r repo.Repo) *Theia {
	return &Theia{s: session, repo: r}
}

func (t *Theia) Send(msg string, editMessage func(string, string) string) (map[string]*discordgo.MessageReference, error) {
	msgRefs := make(map[string]*discordgo.MessageReference, 0)

	allSrvrs, err := t.repo.GetAll()
	if err != nil {
		return nil, err
	}

	for _, v := range allSrvrs {
		srvrMsg := msg
		if editMessage != nil {
			srvrMsg = editMessage(v.GuildID, msg)
			if srvrMsg == "" {
				continue
			}
		}
		ref, err := t.s.ChannelMessageSend(v.ChannelID, srvrMsg)

		if err != nil {
			log.Println("Failed to send message to " + v.GuildID + ":" + v.ChannelID + " - " + err.Error())
			continue
		}

		if ref != nil && ref.Reference() != nil {
			msgRefs[v.GuildID] = ref.Reference()
		}
	}

	return msgRefs, nil
}

func (t *Theia) SendEmbeds(embeds []*discordgo.MessageEmbed, editMessage func(string, []*discordgo.MessageEmbed) []*discordgo.MessageEmbed) (map[string]*discordgo.MessageReference, error) {
	msgRefs := make(map[string]*discordgo.MessageReference, 0)

	allSrvrs, err := t.repo.GetAll()
	if err != nil {
		return nil, err
	}

	for _, v := range allSrvrs {

		srvrEmbeds := embeds
		if editMessage != nil {
			srvrEmbeds = editMessage(v.GuildID, embeds)
			if srvrEmbeds == nil || len(srvrEmbeds) == 0 {
				continue
			}
		}

		ref, err := t.s.ChannelMessageSendEmbeds(v.ChannelID, srvrEmbeds)

		if err != nil {
			log.Println("Failed to send message to " + v.GuildID + ":" + v.ChannelID + " - " + err.Error())
			continue
		}

		if ref != nil && ref.Reference() != nil {
			msgRefs[v.GuildID] = ref.Reference()
		}
	}

	return msgRefs, nil
}

func (t *Theia) SendComplex(msg *discordgo.MessageSend, editMessage func(string, *discordgo.MessageSend) *discordgo.MessageSend) (map[string]*discordgo.MessageReference, error) {
	msgRefs := make(map[string]*discordgo.MessageReference, 0)

	allSrvrs, err := t.repo.GetAll()
	if err != nil {
		return nil, err
	}

	for _, v := range allSrvrs {
		srvrMsg := msg
		if editMessage != nil {
			srvrMsg = editMessage(v.GuildID, msg)
			if srvrMsg == nil {
				continue
			}
		}
		ref, err := t.s.ChannelMessageSendComplex(v.ChannelID, msg)

		if err != nil {
			log.Println("Failed to send message to " + v.GuildID + ":" + v.ChannelID + " - " + err.Error())
			continue
		}

		if ref != nil && ref.Reference() != nil {
			msgRefs[v.GuildID] = ref.Reference()
		}
	}

	return msgRefs, nil
}

func (t *Theia) Inject(name string, cmdDefs []*discordgo.ApplicationCommand, cmdMap map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)) ([]*discordgo.ApplicationCommand, map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	// Init the map with existing values
	injectedMap := make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate), 0)
	injectedList := make([]*discordgo.ApplicationCommand, 0)

	for k, v := range cmdMap {
		injectedMap[k] = v
	}

	for _, v := range cmdDefs {
		injectedList = append(injectedList, v)
	}

	injectedList = append(injectedList, &discordgo.ApplicationCommand{
		Name:        name + "-register",
		Description: "Register this bot to this channel.",
	}, &discordgo.ApplicationCommand{
		Name:        name + "-unregister",
		Description: "Unregister this bot from this server.",
	})

	injectedMap[name+"-register"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guild := i.Interaction.GuildID
		channel := i.Interaction.ChannelID

		err := t.repo.Register(guild, channel)
		log.Println("Registered for " + guild)

		if err != nil {
			log.Println(err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
					Flags:   1 << 6,
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: name + " Bot successfully added to this channel!  \nIf you want to change channel, just run /" + name + "-register in another channel.",
				Flags:   1 << 6,
			},
		})
	}
	injectedMap[name+"-unregister"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guild := i.Interaction.GuildID

		err := t.repo.Unregister(guild)
		log.Println("Unregistered for " + guild)

		if err != nil {
			log.Println(err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
					Flags:   1 << 6,
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: name + " Bot successfully removed from the server.",
				Flags:   1 << 6,
			},
		})
	}

	return injectedList, injectedMap
}
