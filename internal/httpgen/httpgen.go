package httpgen

import (
	"html/template"
	"net/http"
)

var tmpl = `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Header}}</title>
</head>
<body>
    <h1>{{.Header}}</h1>
    {{range .Message}}
        <p>{{.}}</p>
    {{end}}
    <p><a href="{{.WhereToGo}}">Перейти на вкладку погоды</a></p>
</body>
</html>
`

type Data struct {
	Message []string
	Header,
	WhereToGo string
}

func GenerateHttp(w http.ResponseWriter, data *Data) {
	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}
