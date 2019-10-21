package model

type GameType string

const (
	OriginalGame GameType = "original game"
	TranslateGame GameType = "translate game"
	MixedGame GameType = "mixed game"
)

func (gt GameType) String() string { return string(gt) }

type Game struct {
	UserId            int      `json:"userId"`
	Question          string   `json:"question"`
	Answer            string   `json:"answer"`
	CountRightAnswers int      `json:"countRightAnswers"`
	CountAttempts     int      `json:"countAttempts"`
	IsActive          bool     `json:"isActive"`
	Type              GameType `json:"type"`
}
