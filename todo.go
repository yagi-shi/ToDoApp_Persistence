package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3" //ブランクインポート：パッケージの初期化のみを実行し、直接使用しない場合に利用
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
			saveTodos()
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

		saveTodos()

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

		saveTodos()

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

// ファイルからJSONデータを読み込み、todosに変換して格納
func loadTodos() {
	// ①ファイルから読み込み
	data, err := os.ReadFile("todos.json")
	if err != nil {
		if os.IsNotExist(err) { //true->ファイルが存在しない場合のエラー、false->それ以外のエラー(権限エラー等)
			log.Println("初回起動のためデータが存在しません。")
			return
		} else {
			log.Printf("ファイルの読み込みエラー：%v", err)
			return
		}
	}
	fmt.Println("読み込みデータ：", data)

	// ②JSONデータをtodosに変換（デコード）
	err = json.Unmarshal(data, &todos) //todosだと値渡しになり、空のままになるので、&todosで参照渡し(ポインタ渡し)にする必要がある
	if err != nil {
		log.Printf("Json変換エラー：%v", err)
		return
	}
	log.Printf("todos.json から %d 件のデータを読み込みました", len(todos))
}

// データベースを初期化し、テーブルを作る
func initDB() (*sql.DB, error) { //Goの命名規則では、略語は大文字にする（InitDB）
	// ステップ1: データベースファイルを開く
	db, err := sql.Open("sqlite3", "./todos.db") //存在する->開く、存在しない->新規作成
	if err != nil {
		log.Printf("DBオープン失敗：%v", err)

		/*
			Goの？慣習
			エラーが発生した場合、他の戻り値はゼロ値を返す
			- ポインタ → nil
			- 整数 → 0
			- 文字列 → ""
			- bool → false
			- bool → false
		*/
		return nil, err // 成功->db, nil、失敗->nil, errを返してる
	}

	// ステップ2: テーブルを作成
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS todos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL
        )
    `)
	if err != nil {
		log.Printf("テーブル作成失敗：%v", err)
		db.Close() //テーブルが作れなかったらDBを閉じる
		return nil, err
	}

	// ステップ3: DB接続を返す
	return db, nil
}
