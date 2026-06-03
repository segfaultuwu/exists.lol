package bot

import (
	"context"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/auth"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/githubx"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/validate"
	"github.com/segfaultuwu/exists.lol/internal/links"
)

func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "link":
		b.onLink(s, i)
	case "domain":
		b.onDomain(s, i)
	}
}

func (b *Bot) onLink(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := interactionUser(i)
	if user == nil {
		respond(s, i, "❌ Could not detect Discord user.")
		return
	}

	githubUsername := strings.TrimSpace(optionString(i.ApplicationCommandData().Options, "github"))
	if githubUsername == "" {
		respond(s, i, "❌ GitHub username is required.")
		return
	}

	respond(s, i, "⏳ Checking GitHub user `"+githubUsername+"`...")

	exists, err := b.gh.UserExists(context.Background(), githubUsername)
	if err != nil {
		editResponse(s, i, "❌ Failed to check GitHub user:\n```text\n"+err.Error()+"\n```")
		return
	}

	if !exists {
		editResponse(s, i, "❌ GitHub user `"+githubUsername+"` does not exist.")
		return
	}

	err = b.links.Set(links.Link{
		DiscordID:      user.ID,
		DiscordName:    user.Username,
		GitHubUsername: githubUsername,
	})
	if err != nil {
		editResponse(s, i, "❌ Failed to save link:\n```text\n"+err.Error()+"\n```")
		return
	}

	editResponse(s, i, "✅ Linked your Discord account to GitHub `@"+githubUsername+"`.")
}

func (b *Bot) onDomain(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !auth.HasRequiredRole(i.Member, b.cfg.DiscordRequiredRoleID) {
		respond(s, i, "❌ You are not allowed to use this command.")
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 || options[0].Name != "add" {
		respond(s, i, "unknown command")
		return
	}

	user := interactionUser(i)
	if user == nil {
		respond(s, i, "❌ Could not detect Discord user.")
		return
	}

	link, ok, err := b.links.Get(user.ID)
	if err != nil {
		respond(s, i, "❌ Failed to read linked GitHub account.")
		return
	}

	if !ok {
		respond(s, i, "❌ Link your GitHub account first with `/link github:<username>`.")
		return
	}

	subOptions := options[0].Options

	subdomain := optionString(subOptions, "subdomain")
	recordType := strings.ToUpper(optionString(subOptions, "record"))
	value := strings.TrimSpace(optionString(subOptions, "value"))

	if err := validate.Request(subdomain, recordType, value); err != nil {
		respond(s, i, "❌ "+err.Error())
		return
	}

	respond(s, i, "⏳ Creating pull request for `"+subdomain+".exists.lol`...")

	prURL, err := b.gh.CreateDomainPR(context.Background(), githubx.CreateDomainPROptions{
		DiscordUsername: user.Username,
		DiscordID:       user.ID,
		GitHubUsername:  link.GitHubUsername,
		Subdomain:       subdomain,
		RecordType:      recordType,
		Value:           value,
	})
	if err != nil {
		editResponse(s, i, "❌ Failed to create pull request:\n```text\n"+err.Error()+"\n```")
		return
	}

	editResponse(s, i, "✅ Pull request created: "+prURL)
}

func interactionUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}

	if i.User != nil {
		return i.User
	}

	return nil
}

func optionString(options []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, opt := range options {
		if opt.Name == name {
			return opt.StringValue()
		}
	}

	return ""
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println("respond:", err)
	}
}

func editResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		log.Println("edit response:", err)
	}
}
