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
			Description: "Manage exists.lol domain requests",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Request a new exists.lol subdomain",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subdomain",
							Description: "Subdomain name, for example: segfault",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "record",
							Description: "DNS record type",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "CNAME", Value: "CNAME"},
								{Name: "A", Value: "A"},
								{Name: "AAAA", Value: "AAAA"},
								{Name: "TXT", Value: "TXT"},
								{Name: "MX", Value: "MX"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "DNS target, for example: username.github.io",
							Required:    true,
						},
					},
				},
			},
		},
	}
}
