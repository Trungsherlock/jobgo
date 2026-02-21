package skills

// ExtractFromResume extracts known skills from raw resume text.
// Unlike job descriptions, resumes don't have required/preferred sections —
// everything on a resume is a claimed skill.
func ExtractFromResume(text string) []string {
    return findSkills(text) // reuses extractor.go's findSkills
}
