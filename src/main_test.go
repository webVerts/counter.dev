package main

import (
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var app *App

func ResetRedis(){
	user := app.OpenUser("john")
        defer user.Close()
	user.redis.Do("flushdb")
        user.Create("johnjohn")
	user.redis.Do("select", "2")
}

func TestMain(m *testing.M) {

	app = NewApp()
        ResetRedis()
	code := m.Run()
	os.Exit(code)
}

func setupSubTest(t *testing.T) func(t *testing.T) {
    t.Log("setup sub test")
    return func(t *testing.T) {
        t.Log("teardown sub test")
    }
}

func loginCookie(t *testing.T, user string, password string) *apitest.Cookie {
	cookies := apitest.New().
		Handler(app.ServeMux).
		Post("/login").
		FormData("user", user).
		FormData("password", password).
		Expect(t).
		End().
		Response.
		Cookies()

	if len(cookies) == 0 {
		return apitest.NewCookie("swa").Value("")
	}
	return apitest.NewCookie("swa").Value(cookies[0].Value)
}

func TestCreateSuccess(t *testing.T) {
        ResetRedis()
	user := app.OpenUser("peter")
	err := user.Create("mypassmypass")
	assert.Equal(t, err, nil)

}

func TestCreateShortPass(t *testing.T) {
        ResetRedis()
	user := app.OpenUser("peter")
	err := user.Create("mypadd")
	assert.Contains(t, err.Error(), "at least")
}

func TestCreateUsernameTaken(t *testing.T) {
        ResetRedis()
	user := app.OpenUser("peter")
	user.Create("mypassmypass")

	err := user.Create("mypassmypass")
	assert.Contains(t, err.Error(), "Username taken")
}

func TestVerifyPasswordSuccess(t *testing.T) {
        ResetRedis()
	success, _ := app.OpenUser("john").VerifyPassword("johnjohn")
	assert.Equal(t, success, true)
}

func TestVerifyPasswordFail(t *testing.T) {
        ResetRedis()
	success, _ := app.OpenUser("john").VerifyPassword("xxx")
	assert.Equal(t, success, false)
}

func TestApiLoginNoInput(t *testing.T) {
        ResetRedis()
	apitest.New().
		Handler(app.ServeMux).
		Post("/login").
		Expect(t).
		Status(400).
		Body("Missing Input").
		End()
}

func TestApiLoginWrongCredentials(t *testing.T) {
        ResetRedis()
	apitest.New().
		Handler(app.ServeMux).
		Post("/login").
		FormData("user", "xxx").
		FormData("password", "xxx").
		Expect(t).
		Status(400).
		Body("Wrong username or password").
		End()
}

func TestApiLoginSuccess(t *testing.T) {
        ResetRedis()
	apitest.New().
		Handler(app.ServeMux).
		Post("/login").
		FormData("user", "john").
		FormData("password", "johnjohn").
		Expect(t).
		Status(200).
		CookiePresent("swa").
		End()
}

func TestApiAuthSuccess(t *testing.T) {
        ResetRedis()
	apitest.New().
		Handler(app.ServeMux).
		Post("/data").
		Cookies(loginCookie(t, "john", "johnjohn")).
		Expect(t).
		Status(200).
		End()
}

func TestApiAuthFailure(t *testing.T) {
        ResetRedis()
	apitest.New().
		Handler(app.ServeMux).
		Post("/data").
		Cookies(loginCookie(t, "john", "xxx")).
		Expect(t).
		Status(403).
		End()
}


func TestGetPrefEmpty(t *testing.T) {
        ResetRedis()
	user := app.OpenUser("peter")
	err := user.Create("mypassmypass")
	assert.Equal(t, err, nil)
}
