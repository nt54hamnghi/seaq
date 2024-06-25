package config

func (hiku *HikuConfig) Model() string {
	return hiku.GetString("model.name")
}

// func (hiku *HikuConfig) GetAvailableModels() ([]string, error) {
// 	panic("todo")
// }
