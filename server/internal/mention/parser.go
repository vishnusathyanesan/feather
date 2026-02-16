package mention

import (
	"regexp"
	"strings"
)

var mentionRegex = regexp.MustCompile(`@([\w.\-]+)`)

type ParsedMention struct {
	Name string
	Type string // "user_or_group", "channel", "here"
}

func ParseMentions(content string) []ParsedMention {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var mentions []ParsedMention
	for _, match := range matches {
		name := strings.ToLower(match[1])
		if seen[name] {
			continue
		}
		seen[name] = true

		switch name {
		case "channel", "everyone":
			mentions = append(mentions, ParsedMention{Name: name, Type: "channel"})
		case "here":
			mentions = append(mentions, ParsedMention{Name: name, Type: "here"})
		default:
			// Could be a user or group - resolved at service layer
			mentions = append(mentions, ParsedMention{Name: name, Type: "user_or_group"})
		}
	}
	return mentions
}
