package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	ApiKey  string `json:"apikey"`
	IsAdmin int    `json:"isadmin"`
	Enabled int    `json:"enabled"`
}

func (u User) Json() string {
	b, err := json.Marshal(u)
	if err != nil {
		PrintError("User::Json", err)
		return ""
	}
	return string(b)
}

func Json2User(s string) (user User) {
	err := json.Unmarshal([]byte(s), &user)
	if err != nil {
		PrintError("Json2User", err)
		return User{}
	}
	return user
}

func mysqlConnect() *sql.DB {
	conn := GetEnv("zstdfs_mysql", "root:123@tcp(127.0.0.1)/zstdfs")
	db, err := sql.Open("mysql", conn)
	FatalError("ERROR:mysqlConnect.Please set the env var: zstdfs_mysql, it should be: username:password@tcp(your-mysql-ip)/zstdfs", err)
	return db
}

func mysqlPing(db *sql.DB) {
	rows, err := db.Query("SELECT VERSION();")
	defer rows.Close()
	FatalError("mysqlPing", err)
	for rows.Next() {
		var version string
		err = rows.Scan(&version)
		FatalError("mysqlPing", err)
		DebugInfo("mysqlPing", "MySQL version: ", version)
	}
	mysqlInit()
}

func mysqlInit() bool {
	// table users
	q := fmt.Sprintf(`select * from users limit 1;`)
	_, err := sqldb.Prepare(q)
	if err != nil {
		PrintError("mysqlInit:check table users", err)
		//
		table_user, err := embeddedFS.ReadFile("mysql/users.sql")
		if err != nil {
			PrintError("", err)
			return false
		}
		DebugInfo("mysqlInit", string(table_user))
		//
		_, err = sqldb.Exec(string(table_user))
		PrintError("mysqlInit:init table users", err)
		return false
	}

	return true
}

func mysqlUserSignUp(username, password string) bool {
	username = strings.ToLower(username)
	if IsAnyEmpty(username, password) {
		return false
	}
	q := fmt.Sprintf(`insert into users(username,userpass,apikey,isadmin,enabled) values(?,?,?,?,?)`)
	stmt, err := sqldb.Prepare(q)
	if err != nil {
		PrintError("mysqlUserSignUp", err)

		return false
	}

	apikey := GenApiKey(username)

	_, err = stmt.Exec(username, password, apikey, 0, 1)
	if err != nil {
		PrintError("mysqlUserSignUp:stmt.Exec", err)
		return false
	}

	return true
}

func mysqlUserLogin(username, password string, enabled int) (user User) {
	username = strings.ToLower(username)
	if IsAnyEmpty(username, password) {
		return user
	}
	var uname, upass string
	q1 := fmt.Sprintf(`select username,userpass from users where username=? and enabled = ?`)
	stmt1, err := sqldb.Prepare(q1)
	defer stmt1.Close()
	if err != nil {
		return user
	}
	err = stmt1.QueryRow(username, enabled).Scan(&uname, &upass)
	if err != nil {
		return user
	}

	isVerified := VerifyPassword(upass, uname, password)
	DebugInfo("mysqlUserLogin:", username, ":", len(upass), ":", isVerified)
	if !isVerified {
		return user
	}
	//
	q := fmt.Sprintf(`select id,username,apikey,isadmin,enabled from users where username=? and enabled = ?`)
	stmt, err := sqldb.Prepare(q)
	defer stmt.Close()
	if err != nil {
		PrintError("mysqlUserLogin:Prepare", err)
		return user
	}

	row := stmt.QueryRow(username, enabled)
	err = row.Scan(&user.ID, &user.Name, &user.ApiKey, &user.IsAdmin, &user.Enabled)
	if err != nil {
		PrintError("mysqlUserLogin:rows.Scan", err)
		return user
	}

	DebugInfo("mysqlUserLogin:user", user)

	return user

}

func mysqlApiKeyLogin(username, apikey string, enabled int) (user User) {
	username = strings.ToLower(username)
	if IsAnyEmpty(username, apikey) {
		return user
	}
	q := fmt.Sprintf(`select id,username,apikey,isadmin,enabled from users where username=? and apikey=? and enabled = ?`)
	stmt, err := sqldb.Prepare(q)
	if err != nil {
		PrintError("mysqlApiKeyLogin:Prepare", err)
		return user
	}

	row := stmt.QueryRow(username, apikey, enabled)

	DebugInfo("mysqlApiKeyLogin:rows", row)

	err = row.Scan(&user.ID, &user.Name, &user.ApiKey, &user.IsAdmin, &user.Enabled)
	if err != nil {
		PrintError("mysqlApiKeyLogin:rows.Scan", err)
		return user
	}
	stmt.Close()

	DebugInfo("mysqlApiKeyLogin:user", user)

	return user

}
