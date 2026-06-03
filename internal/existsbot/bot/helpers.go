package bot

import (
	"log"

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
