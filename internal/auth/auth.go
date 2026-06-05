package auth

import (
	"slices"

	"github.com/bwmarrin/discordgo"
)

func HasRequiredRole(member *discordgo.Member, requiredRoleID string) bool {
	if requiredRoleID == "" {
		return true
	}

	if member == nil {
		return false
	}

	return slices.Contains(member.Roles, requiredRoleID)
}
