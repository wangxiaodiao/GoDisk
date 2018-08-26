package routers

import (
	"GoDisk/controllers"
	"github.com/astaxie/beego"
)

func init() {
	//页面路由
    beego.Router("/", &controllers.MainController{})
	beego.Router("/admin",&controllers.MainController{},"*:Admin")
    beego.Router("/classify",&controllers.MainController{},"*:Classify")
    beego.Router("/setting",&controllers.MainController{},"*:Setting")
	beego.Router("/postSetting",&controllers.MainController{},"post:PostSetting")
    beego.Router("/localUpload",&controllers.MainController{},"*:LocalUpload")

    //用户模块
	beego.Router("/login",&controllers.UserController{},"*:Login")
	beego.Router("/logout",&controllers.UserController{},"*:Logout")

    //接口Api
    beego.Router("/api/upload",&controllers.ApiController{},"post:Upload")
	beego.Router("/api/saveFile",&controllers.ApiController{},"post:SaveFile")

    //七牛云模块
    beego.Router("/qiniu",&controllers.QiNiuController{},"get:Index") //页面
	beego.Router("/api/qiniu",&controllers.ApiController{},"post:QiniuUpload")//上传接口
}
