package entity

// var ErrPlayerAlreadyExists = errors.New("player already exists")

type Player struct {
	ID     string `json:"id"`
	Mark   string `json:"mark,omitempty"`
	GameID string `json:"game_id,omitempty"`
}

// func (that *Player) setMark() error {
//	if that.Mark == "" {
//		that.Mark = "X"
//
//		return nil
//	}
//
//	return ErrPlayerAlreadyExists
//}
