package section

type App struct {
	Env      string `default:"local"`
	Name     string `default:"test"`
	LogLevel string `default:"trace"`
}
