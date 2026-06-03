package bot

import "github.com/bwmarrin/discordgo"

func Commands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "link",
			Description: "Link your Discord account with your GitHub username",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "github",
					Description: "Your GitHub username",
					Required:    true,
				},
			},
		},
		{
			Name:        "domain",
			Description: "Manage exists.lol domains",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Add a new domain",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subdomain",
							Description: "Subdomain name",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "record",
							Description: "DNS record type",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "A", Value: "A"},
								{Name: "AAAA", Value: "AAAA"},
								{Name: "CNAME", Value: "CNAME"},
								{Name: "TXT", Value: "TXT"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "DNS record value",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "check",
					Description: "Check user's domains",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User to check",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "info",
					Description: "Show info about domain",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "domain",
							Description: "Domain to check",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "self",
			Description: "Manage bot process",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "update",
					Description: "Pull, build and restart the bot",
				},
			},
		},
		{
			Name:        "registry",
			Description: "Manage local domain registry",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "reload",
					Description: "Reload registry from JSON files",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "dump",
					Description: "Show loaded registry domains",
				},
			},
		},
	}
}
