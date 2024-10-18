package entity

const (
	StatusFinished = "finished"
	StatusOngoing  = "ongoing"
	StatusWaiting  = "waiting"
)

type Game struct {
	ID         string    `json:"id"`
	Board      [9]string `json:"board"`
	Winner     string    `json:"winner"`
	Status     string    `json:"status"`
	PlayerTurn string    `json:"player_turn"`
	Players    []*Player `json:"players"`
}

func (that *Game) IsFinished() bool {
	return that.Status == StatusFinished
}

func (that *Game) IsOngoing() bool {
	return that.Status == StatusOngoing
}

func (that *Game) IsWaiting() bool {
	return that.Status == StatusWaiting
}
