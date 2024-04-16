package model

import "github.com/DIMO-Network/model-garage/pkg/vss"

type DIMOData struct {
	ID      string `json:"id"`
	TokenID int    `json:"tokenID"`
	vss.Dimo
}

type SignalCollection struct {
	TokenID uint32 `json:"tokenID"`
}

func (DIMOData) IsNode()            {}
func (this DIMOData) GetID() string { return this.ID }
