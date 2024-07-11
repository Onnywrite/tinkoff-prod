package tokens

import "github.com/Onnywrite/tinkoff-prod/internal/models"

type Pair struct {
	Access  AccessString  `json:"access"`
	Refresh RefreshString `json:"refresh"`
}

func NewPair(usr *models.User) (Pair, error) {
	access := Access{
		Id:    usr.Id,
		Email: usr.Email,
	}
	refresh := Refresh{
		Id: usr.Id,
	}

	accessStr, err := access.Sign()
	if err != nil {
		return Pair{}, err
	}

	refreshStr, err := refresh.Sign()
	if err != nil {
		return Pair{}, err
	}

	return Pair{
		Access:  accessStr,
		Refresh: refreshStr,
	}, nil
}
