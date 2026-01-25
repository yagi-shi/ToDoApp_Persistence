package main

import (
	"log"
	"net/http"
)

func main() {
	// ▫️saveTodosの動作確認
	// todos = append(todos, Todo{ID: 1, Title: "買い物に行く"})
	// todos = append(todos, Todo{ID: 2, Title: "勉強する"})
	// saveTodos()
	// return

	// ▫️loadTodosの動作確認
	// loadTodos()

	// DBを初期化
	db, err := initDB()
	if err != nil {
		log.Fatal("DB初期化失敗:", err)
	}
	defer db.Close()

	err = loadTodosFromDB(db) //initDBのerrはもう使わないので、上書きしてもOK。
	if err != nil {
		log.Printf("データの読み込みに失敗しました。：%v", err)
	}

	http.HandleFunc("/todos", todoHandler)
	http.HandleFunc("/todos/edit", editHandler)
	http.HandleFunc("/todos/update", updateHandler)
	http.HandleFunc("/todos/delete", deleteHandler)
	log.Println("server start :8000")
	http.ListenAndServe(":8000", nil)
}
