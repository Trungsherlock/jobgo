package skills

import (
	"regexp"
	"strings"
)

type JobSkills struct {
	Required []string `json:"required_skills"`
	Preferred []string `json:"preferred_skills"`
	Mentioned []string  `json:"mentioned_skills"`
}

var (
	requiredRe = regexp.MustCompile(`(?i)(requirements?|qualifications?|must.have|what you.?ll need |minimum qualifications?)`)
	preferredRe = regexp.MustCompile(`(?i)(nice.to.have|bonus|preferred qualifications?|what would be (great|nice)|additional qualifications?)`)
)

func ExtractFromJob(description string) JobSkills {
	sections := splitSections(description)
	seen := make(map[string]bool)
	var result JobSkills

	for _, skill := range findSkills(sections["required"]) {
		if !seen[skill] {
			seen[skill] = true
			result.Required = append(result.Required, skill)
		}
	}
	for _, skill := range findSkills(sections["preferred"]) {
		if !seen[skill] {
			seen[skill] = true
			result.Preferred = append(result.Preferred, skill)
		}
	}
	for _, skill := range findSkills(sections["mentioned"]) {
		if !seen[skill] {
			seen[skill] = true
			result.Mentioned = append(result.Mentioned, skill)
		}
	}

	return result
}

func splitSections(text string) map[string]string {
	sections := map[string]string{"required": "", "preferred": "", "mentioned": ""}
	current := "mentioned"
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		if requiredRe.MatchString(t) {
			current = "required"
			continue
		}
		if preferredRe.MatchString(t) {
			current = "preferred"
			continue
		}
		sections[current] += " " + t
	}
	return sections
}

func findSkills(text string) []string  {
	if text == "" {
		return nil
	}
	lower := strings.ToLower(text)
	seen := make(map[string]bool)
	var found []string

	for alias, canonical := range Aliases {
		if strings.Contains(alias, " ") && strings.Contains(lower, alias) && !seen[canonical] {
			seen[canonical] = true
			found = append(found, canonical)
		}
	}
	for alias, canonical := range Aliases {
		if !strings.Contains(alias, " ") && containsWord(lower, alias) && !seen[canonical] {
			seen[canonical] = true
			found = append(found, canonical)
		}
	}
	for _, skill := range Skills {
		if !seen[skill] && containsWord(lower, strings.ToLower(skill)) {
			seen[skill] = true
			found = append(found, skill)
		}
	}

	return found
}

func containsWord(text, word string) bool {
	idx := 0
	for {
		pos := strings.Index(text[idx:], word)
		if pos == -1 {
			return false
		}
		pos += idx
		before := pos == 0 || !isAlphaNum(text[pos-1])
		after := pos+len(word) >= len(text) || !isAlphaNum(text[pos+len(word)])
		if before && after {
			return true
		}
		idx = pos + 1
		if idx >= len(text) {
			return false
		}
	}
}

func isAlphaNum(b byte) bool {
    return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}