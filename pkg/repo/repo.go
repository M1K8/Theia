package repo

type Server struct {
	GuildID   string
	ChannelID string
	Enabled   bool
}

type Repo interface {
	Register(guild, channelID string) error // This takes the ServerID and ChannelID the *-register command is invoked in
	Unregister(guild string) error          // This unregisters the server
	GetAll() ([]Server, error)
	Get(guildID string) (*Server, error)
}
