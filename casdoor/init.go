package casdoor

import (
	_ "embed"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
)

//go:embed token_jwt_key.pem
var JwtPublicKey string

func InitCasdoorConfig() {
	casdoorEndpoint := beego.AppConfig.String("casdoorEndpoint")
	clientId := beego.AppConfig.String("clientId")
	clientSecret := beego.AppConfig.String("clientSecret")
	casdoorOrganization := beego.AppConfig.String("casdoorOrganization")
	casdoorApplication := beego.AppConfig.String("casdoorApplication")
	casdoorsdk.InitConfig(casdoorEndpoint, clientId, clientSecret, JwtPublicKey, casdoorOrganization, casdoorApplication)
}
