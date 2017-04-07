package data

type CardData struct {
	Name string
}

func (data CardData) GetDataName() string {
	return data.Name
}