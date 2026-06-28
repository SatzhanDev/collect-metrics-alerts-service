// Package buildinfo содержит переменные сборки и вспомогательные функции
// для отображения информации о версии приложения.
//
// Переменные заполняются при сборке через -ldflags:
//
//	go build -ldflags "-X github.com/.../internal/buildinfo.Version=v1.0.0" ...
package buildinfo

// Переменные заполняются при сборке через -ldflags.
var (
	Version string
	Date    string
	Commit  string
)

// NA возвращает значение переменной или "N/A" если она пустая.
func NA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
