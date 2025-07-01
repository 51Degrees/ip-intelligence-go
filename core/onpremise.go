package core

type Engine struct {
	dataFile            string
	dataFileUrl         string
	dataFilePullEveryMs int
	isAutoUpdateEnabled bool
}

func InitEngine(defaultDataFileUrl string) *Engine {
	return &Engine{
		dataFileUrl:         defaultDataFileUrl,
		dataFilePullEveryMs: 30 * 60 * 1000, // default 30 minutes
		isAutoUpdateEnabled: true,
	}
}

func (e Engine) IsAutoUpdateEnabled() bool {
	return e.isAutoUpdateEnabled
}
