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
					Name:        "github-pages",
					Description: "Add your subdomain to the github pages",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "subdomain",
							Description: "Subdomain name",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "username",
							Description: "Your github username",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "token",
							Description: "The value github wants you to set, e.g, '2ce6df9a4a5119cbe1927dcb2ba648'",
							Required:    true,
						},
					},
				},
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
					Name:        "redirect",
					Description: "Create a redirect domain",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "domain",
							Description: "Subdomain, for example user or user.exists.lol",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "target",
							Description: "Redirect target URL",
							Required:    true,
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
			ID:                       "",
			ApplicationID:            "",
			GuildID:                  "",
			Version:                  "",
			Type:                     0,
			NameLocalizations:        &map[discordgo.Locale]string{},
			DefaultPermission:        new(bool),
			DefaultMemberPermissions: new(int64),
			NSFW:                     new(bool),
			DMPermission:             new(bool),
			Contexts:                 &[]discordgo.InteractionContextType{},
			IntegrationTypes:         &[]discordgo.ApplicationIntegrationType{},
			DescriptionLocalizations: &map[discordgo.Locale]string{},
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
		{
			Name:        "help",
			Description: "Show all bot commands",
		},
	}
}
