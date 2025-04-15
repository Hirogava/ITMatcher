package models

type AnalyzedResume struct {
	ResumeText  string   `json:"resume_text"`
	Mismatch    []string `json:"mismatch"`
	Coincidence []string `json:"coincidence"`
}

type UserResumeInfo struct {
	ResumeId  int
	Vacancies []struct {
		Name    string
		Percent int
	}
}
