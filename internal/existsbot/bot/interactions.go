package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/segfaultuwu/exists.lol/internal/auth"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/apiclient"
	users "github.com/segfaultuwu/exists.lol/internal/links"
	"github.com/segfaultuwu/exists.lol/internal/validate"
)

func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "link":
		b.onLink(s, i)

	case "domain":
		b.onDomainCommand(s, i)

	case "registry":
		b.onRegistryCommand(s, i)

	case "self":
		b.onSelfCommand(s, i)

	case "help":
		b.onHelp(s, i)

	default:
		respond(s, i, "❌ Unknown command.")
	}
}

func (b *Bot) onHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commands := Commands()

	var out strings.Builder
	out.WriteString("📖 **ExistsBot commands**\n\n")

	for _, cmd := range commands {
		out.WriteString("`/")
		out.WriteString(cmd.Name)
		out.WriteString("`")
		out.WriteString(" — ")
		out.WriteString(cmd.Description)
		out.WriteString("\n")

		for _, opt := range cmd.Options {
			WriteHelpOption(&out, cmd.Name, opt, "")
		}

		out.WriteString("\n")
	}

	respond(s, i, out.String())
}

func (b *Bot) onSelfCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	case "update":
		b.onSelfUpdate(s, i)

	default:
		respond(s, i, "❌ Unknown self subcommand.")
	}
}

func (b *Bot) onSelfUpdate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respond(s, i, "🔄 Starting self-update...")

	script := b.cfg.SelfUpdateScript
	if script == "" {
		script = "./scripts/self-update.sh"
	}

	if _, err := os.Stat(script); err != nil {
		editResponse(s, i, "❌ Update script not found:\n```text\n"+err.Error()+"\n```")
		return
	}

	if err := os.MkdirAll("data", 0o755); err != nil {
		editResponse(s, i, "❌ Failed to create data dir:\n```text\n"+err.Error()+"\n```")
		return
	}

	logPath := filepath.Join("data", "self-update.log")

	logFile, err := os.Create(logPath)
	if err != nil {
		editResponse(s, i, "❌ Failed to create update log:\n```text\n"+err.Error()+"\n```")
		return
	}

	cmd := exec.Command("sh", script)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(
		os.Environ(),
		"SYSTEMD_SERVICE="+b.cfg.SystemdService,
	)

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		editResponse(s, i, "❌ Failed to start updater:\n```text\n"+err.Error()+"\n```")
		return
	}

	editResponse(s, i, fmt.Sprintf(
		"✅ Self-update started.\nPID: `%d`\nLog: `%s`\nCheck with:\n```bash\ncat %s\njournalctl -u %s -f\n```",
		cmd.Process.Pid,
		logPath,
		logPath,
		b.cfg.SystemdService,
	))

	go func() {
		err := cmd.Wait()
		_ = logFile.Close()

		if err != nil {
			log.Println("self-update failed:", err)
			return
		}

		log.Println("self-update finished")
	}()
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

	case "redirect":
		b.onDomainRedirect(s, i, sub)

	case "github-pages":
		b.onDomainGithubPages(s, i, sub)

	default:
		respond(s, i, "❌ Unknown domain subcommand.")
	}
}

func (b *Bot) onDomainGithubPages(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
	user := interactionUser(i)
	if user == nil {
		respond(s, i, "❌ Could not detect Discord user.")
		return
	}

	var subdomain string
	var githubUsername string
	var value string

	for _, opt := range sub.Options {
		switch opt.Name {
		case "subdomain":
			subdomain = strings.TrimSpace(opt.StringValue())
		case "github":
			githubUsername = strings.TrimSpace(opt.StringValue())
		case "value":
			value = strings.TrimSpace(opt.StringValue())
		}
	}

	if subdomain == "" || githubUsername == "" || value == "" {
		respond(s, i, "❌ Missing required options: subdomain, github, value.")
		return
	}

	recordName := strings.Join([]string{
		"_github-pages-challenge-" + githubUsername,
		subdomain,
	}, ".")

	respond(s, i, "⏳ Creating GitHub Pages verification record...")

	res, err := b.api.CreateDomain(context.Background(), apiclient.CreateDomainRequest{
		DiscordID:      user.ID,
		Username:       user.Username,
		GitHubUsername: githubUsername,
		Subdomain:      recordName,
		Records: map[string][]string{
			"TXT": []string{value},
		},
	})
	if err != nil {
		editResponse(s, i, "❌ Failed to create GitHub Pages verification record through local API:\n```text\n"+err.Error()+"\n```")
		return
	}

	editResponse(s, i, fmt.Sprintf(
		"✅ Created GitHub Pages verification record.\n\nRecord:\n`%s`\n\nType:\n`TXT`\n\nValue:\n`%s`",
		res.FQDN,
		value,
	))
}

func (b *Bot) onDomainRedirect(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
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

	domain := strings.TrimSpace(optionString(sub.Options, "domain"))
	target := strings.TrimSpace(optionString(sub.Options, "target"))

	domain = NormalizeSubdomain(domain)

	if domain == "" {
		respond(s, i, "❌ Domain is required.")
		return
	}

	if err := ValidateRedirectTarget(target); err != nil {
		respond(s, i, "❌ "+err.Error())
		return
	}

	if b.cfg.RedirectCNAME == "" {
		respond(s, i, "❌ REDIRECT_CNAME is not configured.")
		return
	}

	respond(s, i, "⏳ Creating redirect `"+domain+".exists.lol`...")

	res, err := b.api.CreateDomain(context.Background(), apiclient.CreateDomainRequest{
		DiscordID:      user.ID,
		Username:       user.Username,
		GitHubUsername: linkedUser.GitHubUsername,
		Subdomain:      domain,
		Records: map[string][]string{
			"CNAME":    []string{b.cfg.RedirectCNAME},
			"REDIRECT": []string{target},
		},
	})
	if err != nil {
		editResponse(s, i, "❌ Failed to create redirect through local API:\n```text\n"+err.Error()+"\n```")
		return
	}

	editResponse(s, i, "✅ Redirect created: `"+res.FQDN+"` → "+target)
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

	input = NormalizeSubdomain(input)

	domain, err := b.api.GetDomain(context.Background(), input)
	if err != nil {
		respond(s, i, "❌ `"+input+".exists.lol` was not found in registry.")
		return
	}

	var out strings.Builder

	out.WriteString("🌐 `")
	out.WriteString(domain.FQDN)
	out.WriteString("`\n\n")

	out.WriteString("**Owner**\n")
	out.WriteString("Username: `")
	out.WriteString(domain.Owner.Username)
	out.WriteString("`\n")

	out.WriteString("GitHub: `@")
	out.WriteString(domain.Owner.GitHub)
	out.WriteString("`\n")

	out.WriteString("Discord: <@")
	out.WriteString(domain.Owner.Discord)
	out.WriteString(">\n\n")

	out.WriteString("**Records**\n")

	for recordType, values := range domain.Records {
		for _, value := range values {
			out.WriteString("• `")
			out.WriteString(recordType)
			out.WriteString(" ")
			out.WriteString(value)
			out.WriteString("`\n")
		}
	}

	respond(s, i, out.String())
}

func (b *Bot) onDomainAdd(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
) {
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

	subdomain = NormalizeSubdomain(subdomain)

	respond(s, i, "⏳ Creating domain `"+subdomain+".exists.lol`...")

	res, err := b.api.CreateDomain(context.Background(), apiclient.CreateDomainRequest{
		DiscordID:      user.ID,
		Username:       user.Username,
		GitHubUsername: linkedUser.GitHubUsername,
		Subdomain:      subdomain,
		Records: map[string][]string{
			recordType: []string{value},
		},
	})
	if err != nil {
		editResponse(s, i, "❌ Failed to create domain through local API:\n```text\n"+err.Error()+"\n```")
		return
	}

	err = b.users.AddDomain(context.Background(), user.ID, users.Domain{
		Subdomain:  subdomain,
		RecordType: recordType,
		Value:      value,
		Status:     "active",
	})
	if err != nil {
		editResponse(s, i, "⚠️ Domain created, but failed to save domain locally:\n```text\n"+err.Error()+"\n```")
		return
	}

	editResponse(s, i, "✅ Domain created: `"+res.FQDN+"`")
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

	all, err := b.api.ListDomains(context.Background())
	if err != nil {
		respond(s, i, "❌ Failed to fetch registry through local API:\n```text\n"+err.Error()+"\n```")
		return
	}

	if len(all) == 0 {
		respond(s, i, "ℹ️ Registry is empty.")
		return
	}

	lines := make([]string, 0, len(all))

	for _, domain := range all {
		var line strings.Builder

		line.WriteString(domain.FQDN)
		line.WriteString(" | discord_id=")
		line.WriteString(domain.Owner.Discord)
		line.WriteString(" | github=")
		line.WriteString(domain.Owner.GitHub)

		lines = append(lines, line.String())
	}

	chunks := SplitDiscordCodeBlock("📦 Loaded registry domains:", lines)

	respond(s, i, chunks[0])

	for _, chunk := range chunks[1:] {
		Followup(s, i, chunk)
	}
}

func (b *Bot) onRegistryReload(s *discordgo.Session, i *discordgo.InteractionCreate) {
	res, err := b.api.ReloadRegistry(context.Background())
	if err != nil {
		respond(s, i, "❌ Failed to reload registry through local API:\n```text\n"+err.Error()+"\n```")
		return
	}

	respond(s, i, fmt.Sprintf("✅ Registry reloaded. Loaded `%d` domains.", res.Domains))
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

	all, err := b.api.ListDomains(context.Background())
	if err != nil {
		respond(s, i, "❌ Failed to fetch domains through local API:\n```text\n"+err.Error()+"\n```")
		return
	}

	domains := make([]apiclient.DomainResponse, 0)

	for _, domain := range all {
		if domain.Owner.Discord == target.ID {
			domains = append(domains, domain)
		}
	}

	if len(domains) == 0 {
		respond(s, i, "ℹ️ This user has no domains.")
		return
	}

	var out strings.Builder

	out.WriteString("🌐 Domains for <@")
	out.WriteString(target.ID)
	out.WriteString(">:\n\n")

	for _, domain := range domains {
		out.WriteString("• `")
		out.WriteString(domain.FQDN)
		out.WriteString("`\n")

		for recordType, values := range domain.Records {
			for _, value := range values {
				out.WriteString("  • `")
				out.WriteString(recordType)
				out.WriteString(" ")
				out.WriteString(value)
				out.WriteString("`\n")
			}
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

	if b.users.IsAlreadyLinked(githubUsername, user.ID) {
		respond(s, i, "You've already linked this account")
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
