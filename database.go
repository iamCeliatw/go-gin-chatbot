// package main

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"log"
// 	"net/http"

// 	"github.com/gorilla/mux" // 这是一个常用的路由库
// 	_ "github.com/lib/pq"
// )

// var db *sql.DB

// type User struct {
// 	ID    int    `json:"id"`
// 	Name  string `json:"name"`
// 	Email string `json:"email"`
// }

// func initDB() {
// 	var err error
// 	db, err = sql.Open("postgres", "user=postgres password=ar0708 dbname=postgres sslmode=disable host=localhost port=5434")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	err = db.Ping()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// // create使用者
// func createUser(w http.ResponseWriter, r *http.Request) {
// 	// 解析表单数据
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// 获取表单字段
// 	name := r.FormValue("name")
// 	email := r.FormValue("email")

// 	// 执行数据库插入操作
// 	_, err := db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", name, email)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode("User created")
// }

// func getUser(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)
// 	id := params["id"]

// 	var user User
// 	row := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id)
// 	err := row.Scan(&user.ID, &user.Name, &user.Email)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(user)
// }

// func updateUser(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)
// 	id := params["id"]

// 	// 解析表单数据
// 	if err := r.ParseForm(); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// 获取表单字段
// 	name := r.FormValue("name")
// 	email := r.FormValue("email")

// 	// 执行数据库更新操作
// 	_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", name, email, id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode("User updated")
// }

// func deleteUser(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)
// 	id := params["id"]

// 	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode("User deleted")
// }

// func main() {
// 	initDB()
// 	defer db.Close()

// 	r := mux.NewRouter()
// 	r.HandleFunc("/user", createUser).Methods("POST")
// 	r.HandleFunc("/user/{id}", getUser).Methods("GET")
// 	r.HandleFunc("/user/{id}", updateUser).Methods("PUT")
// 	r.HandleFunc("/user/{id}", deleteUser).Methods("DELETE")

// 	log.Fatal(http.ListenAndServe(":8080", r))
// }
