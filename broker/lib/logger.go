package lib

type Logger interface {
	Error(string, string, error)
	Info(string, string)
}
