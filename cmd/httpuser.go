package cmd

import (
	"strings"

	"github.com/kataras/iris/v12"
)

func getCurrentUser(ctx iris.Context) (user User) {
	curUser := ctx.GetCookie("currentUser")
	if curUser == "" {
		ctx.Redirect("/signin", 302)
		return
	}
	//DebugInfo("getCurrentUser:curUser", curUser)
	user = JSON2User(Decrypt(curUser))
	if user.Name != "" && user.Enabled != 0 {
		return user
	}
	DebugInfo("getCurrentUser:OK", user.Name)

	ctx.Redirect("/signin", 302)
	return user
}

func signupIndex(ctx iris.Context) {
	ctx.View("login.html", iris.Map{
		"title":      "New User Signup:",
		"frm_action": "/usersignup",
		"frm_hash":   getFormHash("new"),
		"frm_submit": "Sign Up",
	})
}

func userSignup(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("userSignup", "pls use POST")
		return
	}
	username := ctx.PostValue("username")
	password := ctx.PostValue("password")
	frmhash := ctx.PostValue("frm_hash")
	if IsAnyEmpty(username, password, frmhash) {
		DebugInfo("userSignup", username, ":", password, ":", frmhash)
		return
	}
	if verifyFormHash(frmhash) == false {
		DebugInfo("userSignup", "frm_hash is invalid")
		return
	}
	passwordhash := GetPassword(username, password)
	DebugInfo("userSignup:password", len(password), ":passwordhash:", passwordhash, ":", len(passwordhash))
	success := mysqlUserSignUp(username, passwordhash)
	if success {
		mongoUserCollectionInit(username)
		mongoAdminCreateIndex(username)
		ctx.Redirect("/signin", 302)
	} else {
		ctx.Redirect("/signup", 302)
	}
}

func loginIndex(ctx iris.Context) {
	ctx.View("login.html", iris.Map{
		"title":      "User Signin:",
		"frm_action": "/userlogin",
		"frm_hash":   getFormHash("login"),
		"frm_submit": "Login",
	})
}

func userLogin(ctx iris.Context) {
	if ctx.Method() != "POST" {
		DebugInfo("userLogin", "pls use POST")
		return
	}
	username := strings.ToLower(ctx.PostValue("username"))
	password := ctx.PostValue("password")
	frmhash := ctx.PostValue("frm_hash")

	if IsAnyEmpty(username, password, frmhash) {
		DebugInfo("userLogin", username, ":", frmhash)
		return
	}
	if verifyFormHash(frmhash) == false {
		DebugInfo("userLogin", "frmhash is invalid")
		return
	}

	user := mysqlUserLogin(username, password, 1)
	DebugInfo("userLogin:user", user, ":", user.JSON())
	//
	cookieuser := Encrypt(user.JSON())
	DebugInfo("userLogin:enc:cookieuser", cookieuser)
	DebugInfo("userLogin:dec:cookieuser", Decrypt(cookieuser))

	if user.Enabled == 1 && user.Name != "" {
		ctx.SetCookieKV("currentUser", cookieuser)
		ctx.Redirect("/home", 302)
	} else {
		ctx.Redirect("/signin", 302)
	}

}

func logoutIndex(ctx iris.Context) {
	ctx.RemoveCookie("currentUser")
	ctx.Redirect("/", 302)
}

func getFormHash(salt string) string {
	k1 := UnixFormat(GetNowUnix(), "2006-01-02")
	k2 := strings.Join([]string{SHA256String(k1), salt}, ":")
	k3 := GetXxhashString([]byte(k2))
	return strings.Join([]string{k3, salt}, ":")
}

func verifyFormHash(s string) bool {
	if strings.Index(s, ":") < 1 {
		return false
	}
	hashsalt := strings.Split(s, ":")
	if len(hashsalt) != 2 {
		return false
	}

	k1 := UnixFormat(GetNowUnix(), "2006-01-02")
	k2 := strings.Join([]string{SHA256String(k1), hashsalt[1]}, ":")
	k3 := GetXxhashString([]byte(k2))
	if hashsalt[0] == k3 {
		return true
	}
	return false
}
