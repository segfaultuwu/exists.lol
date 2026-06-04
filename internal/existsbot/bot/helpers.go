package bot

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const discordMessageLimit = 2000

func SplitDiscordCodeBlock(header string, lines []string) []string {
	const limit = 1900
	var chunks []string

	current := header + "\n\n```text\n"

	for _, line := range lines {
		if len(current)+len(line)+1+3 > limit {
			current += "```"
			chunks = append(chunks, current)

			current = "```text\n"
		}

		current += line + "\n"
	}

	current += "```"
	chunks = append(chunks, current)

	return chunks
}

func Followup(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Println("followup:", err)
	}
}

func WriteHelpOption(
	out *strings.Builder,
	commandName string,
	opt *discordgo.ApplicationCommandOption,
	parent string,
) {
	switch opt.Type {
	case discordgo.ApplicationCommandOptionSubCommand:
		out.WriteString("  `/")
		out.WriteString(commandName)
		out.WriteString(" ")

		if parent != "" {
			out.WriteString(parent)
			out.WriteString(" ")
		}

		out.WriteString(opt.Name)

		for _, subOpt := range opt.Options {
			if subOpt.Type == discordgo.ApplicationCommandOptionSubCommand {
				continue
			}

			out.WriteString(" ")

			if subOpt.Required {
				out.WriteString("<")
				out.WriteString(subOpt.Name)
				out.WriteString(">")
			} else {
				out.WriteString("[")
				out.WriteString(subOpt.Name)
				out.WriteString("]")
			}
		}

		out.WriteString("` — ")
		out.WriteString(opt.Description)
		out.WriteString("\n")

		for _, subOpt := range opt.Options {
			if subOpt.Type == discordgo.ApplicationCommandOptionSubCommand {
				WriteHelpOption(out, commandName, subOpt, opt.Name)
			}
		}

	case discordgo.ApplicationCommandOptionSubCommandGroup:
		for _, subOpt := range opt.Options {
			WriteHelpOption(out, commandName, subOpt, opt.Name)
		}
	}
}

func NormalizeSubdomain(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimSuffix(input, ".")
	input = strings.TrimSuffix(input, ".exists.lol")
	input = strings.ToLower(input)

	return input
}

func ValidateRedirectTarget(target string) error {
	if target == "" {
		return fmt.Errorf("redirect target is required")
	}

	u, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("invalid redirect target")
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("redirect target must start with http:// or https://")
	}

	if u.Host == "" {
		return fmt.Errorf("redirect target must include host")
	}

	return nil
}
