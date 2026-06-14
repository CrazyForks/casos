package controllers

import (
	"os"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
)

func (c *ApiController) Signin() {
	code := c.Input().Get("code")
	state := c.Input().Get("state")

	token, err := casdoorsdk.GetOAuthToken(code, state)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	claims, err := casdoorsdk.ParseJwtToken(token.AccessToken)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	claims.AccessToken = token.AccessToken
	c.SetSessionClaims(claims)

	c.ResponseOk(claims)
}

func (c *ApiController) Signout() {
	c.SetSessionClaims(nil)

	c.ResponseOk()
}

func (c *ApiController) GetAccount() {
	if c.RequireSignedIn() {
		return
	}

	claims := c.GetSessionClaims()
	hostname, _ := os.Hostname()

	c.ResponseOk(claims, hostname)
}
