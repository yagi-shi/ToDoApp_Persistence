package main

import (
	"html/template"
	"net/http"
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

	tmpl := template.Must(template.ParseFiles("templates/index.html")) //index.htmlを読み込んでtmplに格納（＝準備）
	tmpl.Execute(w, todos)                                             //データを埋め込んでレスポンスとして返す（＝実行）
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// クエリパラメータからIDを取得
		idStr := r.URL.Query().Get("id")

		// 取得したIDを整数に変換
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
				found = true // 見つかった場合、フラグをtrueに設定
				break        // 見つかった段階でループを抜けることで、無駄な処理を避ける
			}
		}
		if !found {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}

		// 4. テンプレート実行
		tmpl := template.Must(template.ParseFiles("templates/edit.html"))
		tmpl.Execute(w, targetTodo)
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		idStr := r.URL.Query().Get("id")
		newTitle := r.FormValue("title")

		// IDを整数に変換
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// 対応するTodoを更新
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

		// IDを整数に変換
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// 対応するTodoを削除
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
