package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/auth"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/githubx"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/validate"
	users "github.com/segfaultuwu/exists.lol/internal/links"
)

func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "link":
		b.onLink(s, i)

	case "domain":
		b.onDomainCommand(s, i)

	case "registry":
		b.onRegistryCommand(s, i)

	default:
		respond(s, i, "❌ Unknown command.")
	}
}

func (b *Bot) onDomainCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	if len(data.Options) == 0 {
		respond(s, i, "❌ Missing subcommand.")
		return
	}

	sub := data.Options[0]

	switch sub.Name {
	case "add":
		b.onDomainAdd(s, i, sub)

	case "check":
		b.onDomainCheck(s, i, sub)

	case "info":
		b.onDomainInfo(s, i, sub)

	default:
		respond(s, i, "❌ Unknown domain subcommand.")
	}
}

func (b *Bot) onDomainInfo(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	input := strings.TrimSpace(optionString(sub.Options, "domain"))
	if input == "" {
		respond(s, i, "❌ Domain is required.")
		return
	}

	input = strings.TrimSuffix(input, ".")
	input = strings.TrimSuffix(input, ".exists.lol")
	input = strings.ToLower(input)

	domain, ok := b.registry.Get(input)
	if !ok {
		respond(s, i, "❌ `"+input+".exists.lol` was not found in registry.")
		return
	}

	var out strings.Builder

	out.WriteString("🌐 `")
	out.WriteString(input)
	out.WriteString(".exists.lol`\n\n")

	out.WriteString("**Owner**\n")
	out.WriteString("Username: `")
	out.WriteString(domain.Owner.Username)
	out.WriteString("`\n")

	out.WriteString("GitHub: `@")
	out.WriteString(domain.Owner.GitHubUsername)
	out.WriteString("`\n")

	out.WriteString("Discord: <@")
	out.WriteString(domain.Owner.DiscordID)
	out.WriteString(">\n\n")

	out.WriteString("**Records**\n")

	for recordType, value := range domain.Records {
		out.WriteString("• `")
		out.WriteString(recordType)
		out.WriteString(" ")
		out.WriteString(value)
		out.WriteString("`\n")
	}

	respond(s, i, out.String())
}

func (b *Bot) onDomainAdd(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	if !auth.HasRequiredRole(i.Member, b.cfg.DiscordRequiredRoleID) {
		respond(s, i, "❌ You are not allowed to use this command.")
		return
	}

	user := interactionUser(i)
	if user == nil {
		respond(s, i, "❌ Could not detect Discord user.")
		return
	}

	linkedUser, ok, err := b.users.GetByDiscordID(user.ID)
	if err != nil {
		respond(s, i, "❌ Failed to read linked GitHub account.")
		return
	}

	if !ok {
		respond(s, i, "❌ Link your GitHub account first with `/link github:<username>`.")
		return
	}

	subdomain := strings.TrimSpace(optionString(sub.Options, "subdomain"))
	recordType := strings.ToUpper(strings.TrimSpace(optionString(sub.Options, "record")))
	value := strings.TrimSpace(optionString(sub.Options, "value"))

	if err := validate.Request(subdomain, recordType, value); err != nil {
		respond(s, i, "❌ "+err.Error())
		return
	}

	respond(s, i, "⏳ Creating pull request for `"+subdomain+".exists.lol`...")

	prURL, err := b.gh.CreateDomainPR(context.Background(), githubx.CreateDomainPROptions{
		DiscordUsername: user.Username,
		DiscordID:       user.ID,
		GitHubUsername:  linkedUser.GitHubUsername,
		Subdomain:       subdomain,
		RecordType:      recordType,
		Value:           value,
	})
	if err != nil {
		editResponse(s, i, "❌ Failed to create pull request:\n```text\n"+err.Error()+"\n```")
		return
	}

	err = b.users.AddDomain(context.Background(), user.ID, users.Domain{
		Subdomain:  subdomain,
		RecordType: recordType,
		Value:      value,
		Status:     "pending",
	})
	if err != nil {
		editResponse(s, i, "⚠️ Pull request created, but failed to save domain locally:\n"+prURL+"\n```text\n"+err.Error()+"\n```")
		return
	}

	editResponse(s, i, "✅ Pull request created: "+prURL)
}
func (b *Bot) onRegistryCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !auth.HasRequiredRole(i.Member, b.cfg.DiscordRequiredRoleID) {
		respond(s, i, "❌ You are not allowed to use this command.")
		return
	}

	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		respond(s, i, "❌ Missing subcommand.")
		return
	}

	switch data.Options[0].Name {
	case "reload":
		b.onRegistryReload(s, i)

	case "dump":
		b.onRegistryDump(s, i)

	default:
		respond(s, i, "❌ Unknown registry subcommand.")
	}
}

func (b *Bot) onRegistryDump(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !auth.HasRequiredRole(i.Member, b.cfg.DiscordRequiredRoleID) {
		respond(s, i, "❌ You are not allowed to use this command.")
		return
	}

	all := b.registry.All()

	if len(all) == 0 {
		respond(s, i, "ℹ️ Registry is empty.")
		return
	}

	lines := make([]string, 0, len(all))

	for subdomain, domain := range all {
		var line strings.Builder

		line.WriteString(subdomain)
		line.WriteString(".exists.lol")
		line.WriteString(" | discord_id=")
		line.WriteString(domain.Owner.DiscordID)
		line.WriteString(" | github=")
		line.WriteString(domain.Owner.GitHubUsername)

		lines = append(lines, line.String())
	}

	chunks := SplitDiscordCodeBlock("📦 Loaded registry domains:", lines)

	respond(s, i, chunks[0])

	for _, chunk := range chunks[1:] {
		Followup(s, i, chunk)
	}
}

func (b *Bot) onRegistryReload(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := b.registry.Reload(b.cfg.RegistryDir); err != nil {
		respond(s, i, "❌ Failed to reload registry:\n```text\n"+err.Error()+"\n```")
		return
	}

	count := len(b.registry.All())

	respond(s, i, fmt.Sprintf("✅ Registry reloaded. Loaded `%d` domains.", count))
}

func (b *Bot) onDomainCheck(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	target := interactionUser(i)

	for _, opt := range sub.Options {
		if opt.Name == "user" {
			target = opt.UserValue(s)
		}
	}

	if target == nil {
		respond(s, i, "❌ Could not detect user.")
		return
	}

	domains := b.registry.ByDiscordID(target.ID)

	if len(domains) == 0 {
		errors := b.registry.LastErrors()

		if len(errors) > 0 {
			respond(s, i, "ℹ️ This user has no domains.\n\n⚠️ Registry has skipped files. Use `/registry reload` to see errors.")
			return
		}

		respond(s, i, "ℹ️ This user has no domains.")
		return
	}

	var out strings.Builder

	out.WriteString("🌐 Domains for <@")
	out.WriteString(target.ID)
	out.WriteString(">:\n\n")

	for subdomain, domain := range domains {
		out.WriteString("• `")
		out.WriteString(subdomain)
		out.WriteString(".exists.lol`")

		for recordType, value := range domain.Records {
			out.WriteString(" → `")
			out.WriteString(recordType)
			out.WriteString(" ")
			out.WriteString(value)
			out.WriteString("`")
		}

		out.WriteString("\n")
	}

	respond(s, i, out.String())
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

	err = b.users.Link(users.User{
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

func optionRequiredString(options []*discordgo.ApplicationCommandInteractionDataOption, name string) (string, error) {
	value := strings.TrimSpace(optionString(options, name))
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}

	return value, nil
}
