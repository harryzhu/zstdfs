package cmd

import (
	"strings"

	"github.com/kataras/iris/v12"
)

func userLogout(ctx iris.Context) {
	ctx.Logout()
	ctx.View("logout.html", iris.Map{
		"send":     "/admin/edit",
		"redirect": "/",
	})
}

func userAdd(ctx iris.Context) {
	useraddName := ctx.PostValue("useraddname")
	useraddPassword := ctx.PostValue("useraddpassword")
	if IsAnyEmpty(useraddName, useraddPassword) {
		DebugWarn("userAdd", "name and password cannot be empty")
		return
	}

	if strings.HasPrefix(useraddName, "_") || strings.HasPrefix(useraddName, ".") {
		ctx.Writef("ERROR:userAdd: %s", "name could not be start with . or _")
		return
	}

	err := dbUserAdd(useraddName, useraddPassword)
	if err != nil {
		ctx.Writef("ERROR:userAdd: %s", err.Error())
	} else {
		updateUserPass()
		ctx.Redirect("/")
	}
}

func userDelete(ctx iris.Context) {
	userdeleteName := ctx.PostValue("userdeletename")
	if IsAnyEmpty(userdeleteName) {
		DebugWarn("userDelete", "name cannot be empty")
		return
	}

	err := dbUserDelete(userdeleteName)
	if err != nil {
		ctx.Writef("ERROR:userDelete: %s", err.Error())
	} else {
		updateUserPass()
		ctx.Redirect("/")
	}

}

func userSignup(ctx iris.Context) {
	ctx.View("signup.html")
}
