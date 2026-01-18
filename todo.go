package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Todo struct {
	ID    int
	Title string
}

var todos []Todo

func todoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		title := r.FormValue("title")
		if title != "" {
			todos = append(todos, Todo{ID: getNextID(), Title: title})
		}
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, todos)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		idStr := r.URL.Query().Get("id")

		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var targetTodo Todo
		var found bool
		for _, todo := range todos {
			if todo.ID == idInt {
				targetTodo = todo
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/edit.html"))
		tmpl.Execute(w, targetTodo)
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		idStr := r.URL.Query().Get("id")
		newTitle := r.FormValue("title")

		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var found bool
		for i, todo := range todos {
			if todo.ID == idInt {
				todos[i].Title = newTitle
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}

		// PRGパターン: POST後はリダイレクトし、リロードによる重複送信を防ぐ
		http.Redirect(w, r, "/todos", http.StatusSeeOther)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		idStr := r.URL.Query().Get("id")

		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var found bool
		for i, todo := range todos {
			if todo.ID == idInt {
				todos = append(todos[:i], todos[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, "/todos", http.StatusSeeOther)
	}
}

func getNextID() int {
	maxId := 0
	for _, todo := range todos {
		if todo.ID > maxId {
			maxId = todo.ID
		}
	}
	return maxId + 1
}

// todosをJSONに変換し、ファイルに保存
func saveTodos() {
	//①Jsonデータに変換（エンコード）
	data, err := json.MarshalIndent(todos, "", " ") // Marshal:軽いが可読性低／MarshalIndent:重いが可読性高
	if err != nil {
		// log:stderr（標準エラー）、タイムスタンプ有
		// fmt:標準エラー（stderr）、タイムスタンプ無
		log.Printf("Json変換エラー：%v", err) //log.Fatall(err)は、強制終了する挙動
		return
	}
	fmt.Println("保存処理：", data)

	// ②ファイルに書き込み

	//ここで書き込みを実施してて、err有->err、err無->nilになり、下記処理に続く。だから、成功だった場合というの処理は不要。
	err = os.WriteFile("todos.json", data, 0644)
	if err != nil {
		log.Printf("ファイルの書き込みエラー：%v", err)
		return
	}
}
