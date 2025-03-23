package config

const (
	appName = "APP_NAME"
	appEnv  = "APP_ENV"
	appLang = "APP_LANG"
	appLog  = "APP_LOG"
)

type AppConfig interface {
	Name() string
	Env() string
	Lang() string
	Log() bool
}

type appConfig struct {
	name string
	env  string
	lang string
	log  bool
}

func NewAppConfig() (AppConfig, error) {
	name := getEnv(appName, "ct")
	env := getEnv(appEnv, "dev")
	lang := getEnv(appLang, "ru")
	log := getEnv(appLog, "false")

	return &appConfig{
		name: name,
		env:  env,
		lang: lang,
		log:  log == "true",
	}, nil
}

func (cfg *appConfig) Name() string {
	return cfg.name
}

func (cfg *appConfig) Env() string {
	return cfg.env
}

func (cfg *appConfig) Lang() string {
	return cfg.lang
}

func (cfg *appConfig) Log() bool {
	return cfg.log
}
