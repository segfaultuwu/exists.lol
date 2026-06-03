package auth

import "github.com/bwmarrin/discordgo"

func HasRequiredRole(member *discordgo.Member, requiredRoleID string) bool {
	if requiredRoleID == "" {
		return true
	}

	if member == nil {
		return false
	}

	for _, roleID := range member.Roles {
		if roleID == requiredRoleID {
			return true
		}
	}

	return false
}
