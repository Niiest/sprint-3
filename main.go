package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/go-chi/chi/v5"
)

// Структура для хранения информации о работнике
type Employee struct {
	FullName   string
	Department string
	Salaries   []float64
}

// Шаблон для отображения информации о работниках
const employeeTemplate = `{{range .}}` +
	`Работник: {{.FullName}} {{if .Department}}/ Отдел: {{.Department}}{{end}}
 Выплаченная зарплата: {{range .Salaries}}{{.}} {{end}}
 {{end}}`

func main() {
	r := chi.NewRouter()
	r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("chi"))
	})
	r.Get("/metrics", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("chi"))
	})
	r.Get("/item/{id}", func(rw http.ResponseWriter, r *http.Request) {
		// получаем значение URL-параметра id
		id := chi.URLParam(r, "id")
		io.WriteString(rw, fmt.Sprintf("item = %s", id))
	})

	// Данные о работниках
	employees := []Employee{
		{
			FullName: "Александр Смирнов",
			Salaries: []float64{50000, 52000, 54000, 55000},
		},
		{
			FullName:   "Екатерина Петрова",
			Department: "Маркетинг",
			Salaries:   []float64{60000, 62000, 64000, 65000},
		},
	}

	// Создание и парсинг шаблона
	tmpl, err := template.New("employeeList").Parse(employeeTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при парсинге шаблона: %v\n", err)
		os.Exit(1)
	}

	// Выполнение шаблона и вывод результата
	if err := tmpl.Execute(os.Stdout, employees); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при выполнении шаблона: %v\n", err)
		os.Exit(1)
	}

	// r передаётся как http.Handler
	http.ListenAndServe(":8084", r)
}
