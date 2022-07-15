### Summary
Theia allows for simple multiplexing of messages for Discord bots written using [discordgo](https://github.com/bwmarrin/discordgo).

Flow:

- Call `NewTheia(*discordgo.Session, repo.Repo)` to intantiate the multiplexer
- Call `theia.Inject(bot_name, []*discordgo.ApplicationCommand, map)` on you Commands Map. This will create a `{name}-register` and `{name}-unregister` slash commands for server owners to use.
- To send a message through the multiplexer, use `theia.Send(string, func)`, `theia.SendEmbeds([]*discordgo.Embeds, func)`, or `theia.SendComplex(*discordgo.MessageSend, func)`.
    - The 2nd `func` parameter will be run on each message for each server, and can be used to customize / filter messages for a specific server. Passing `nil` here will send a message as normal.
    - If an error is encountered, it will be logged & Theia will continue onto the next server.
    - All of these functions return a map[string]*discordgo.MessageReference that can be used to edit / remove the sent message. The map is index based on the GuildID.


#### Example
[Here](https://github.com/M1K8/Theia/blob/main/cmd/example/main.go)