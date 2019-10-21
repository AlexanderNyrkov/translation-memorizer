package common

type Commands string

const (
	ShowAllWords Commands = "show_me_all_words"
	StartGame Commands    = "start_game"
	StopGame Commands     = "stop_game"
	DeleteWord Commands   = "delete_word"
	Help Commands		  = "help"
)

func (c Commands) String() string { return string(c) }

