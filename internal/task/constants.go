package task

const (
	StatusNew     = "new"
	StatusRunning = "in_progress"
	StatusStopped = "done"
)

var statusMap = map[string]interface{}{
	StatusNew:     nil,
	StatusRunning: nil,
	StatusStopped: nil,
}

const (
	IdMinLen          = 1
	IdMaxLen          = 64
	TitleMinLen       = 3
	TitleMaxLen       = 100
	DescriptionMinLen = 0
	DescriptionMaxLen = 500
)
