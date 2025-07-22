package core

type GroupConfig struct {
	Configs []*Config
}

type Group struct {
	clientMap map[string]*Client
}

func NewGroup(groupConfig *GroupConfig) *Group {
	var group = Group{clientMap: make(map[string]*Client)}
	var configs = groupConfig.Configs
	for _, cfg := range configs {
		if group.clientMap[cfg.Key] != nil {
			panic("key is exist " + cfg.Key)
		}
		var client = NewClient(*cfg)
		group.clientMap[cfg.Key] = client
	}
	return &group
}

func (g *Group) Get(key string) *Client {
	return g.clientMap[key]
}
