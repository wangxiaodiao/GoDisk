/***********************

	api接口

************************/

package controllers

import (
	"GoDisk/models"
	"GoDisk/tools"
	"encoding/base64"
	"encoding/json"
	"github.com/astaxie/beego"
	"log"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ApiController struct {
	beego.Controller
}

// 添加分类api  路由 /api/category/add
func (this *ApiController) CategoryAdd() {
	name := this.GetString("name")
	key := this.GetString("key")
	description := this.GetString("description")
	info := &models.Category{Name: name, Key: key, Description: description}
	err := models.AddCategory(info)
	var data *ResultData
	if err != nil {
		data = &ResultData{Error: 1, Title: "失败:", Msg: "添加失败！"}
	} else {
		data = &ResultData{Error: 0, Title: "成功:", Msg: "添加成功！"}
	}
	this.Data["json"] = data
	this.ServeJSON()
}

// 修改分类api  路由 /api/category/update
func (this *ApiController) CategoryUpdate() {
	id := this.GetString("id")
	data := &models.Category{}
	info := &ResultData{}
	if err := this.ParseForm(data); err != nil {
		info = &ResultData{Error: 1, Title: "失败:", Msg: "接收表单数据出错！"}
	} else {
		data.Id = tools.StringToInt(id)
		err := models.UpdateCategory(data)
		if err != nil {
			info = &ResultData{Error: 1, Title: "失败:", Msg: "数据库操作出错！"}
		} else {
			info = &ResultData{Error: 0, Title: "成功:", Msg: "修改成功！"}
		}
	}
	this.Data["json"] = info
	this.ServeJSON()
}

// 删除分类api  路由 /api/category/delete
func (this *ApiController) CategoryDelete() {
	info := &ResultData{}
	//先判断分类数目 为1时，不允许删除
	count, _ := models.TableNumber("category")
	if count == 1 {
		info = &ResultData{Error: 1, Title: "失败:", Msg: "必须保留一个分类！"}
	} else {
		id, _ := strconv.Atoi(this.GetString("id"))
		err := models.DeleteCategory(id)
		if err != nil {
			info = &ResultData{Error: 1, Title: "失败:", Msg: "数据库操作出错！"}
		} else {
			info = &ResultData{Error: 0, Title: "成功:", Msg: "删除成功！"}
		}
	}
	this.Data["json"] = info
	this.ServeJSON()
}

// 分类列表 路由 /api/category/list
func (this *ApiController) CategoryList() {
	this.Data["json"] = &JsonData{Code: 0, Count: 100, Msg: "", Data: models.GetCategoryJson()}
	this.ServeJSON()
}

// 网站设置页面  路由  /api/site/config
func (this *ApiController) SiteConfig() {
	// 判断提交类型 user为用户信息表单  site为网站配置表单
	submit := this.GetString("submit")
	info := &ResultData{}
	var data interface{}
	if submit == "userInfo" {
		data = &models.UserConfigOption{}
	} else if submit == "siteInfo" {
		data = &models.SiteConfigOption{}
	} else {
		data = &models.QiniuConfigOption{}
	}
	if err := this.ParseForm(data); err != nil {
		info = &ResultData{Error: 1, Title: "失败:", Msg: "接收表单数据出错！"}
	} else {
		t := reflect.TypeOf(data).Elem()  //类型
		v := reflect.ValueOf(data).Elem() //值
		for i := 0; i < t.NumField(); i++ {
			config := &models.Config{Option: t.Field(i).Name, Value: v.Field(i).String()}
			err := models.SiteConfig(config)
			if err != nil {
				info = &ResultData{Error: 1, Title: "失败:", Msg: "出现未知错误！"}
			} else {
				info = &ResultData{Error: 0, Title: "成功:", Msg: "信息重置成功！"}
			}
		}
	}
	this.Data["json"] = info
	this.ServeJSON()
}

// 文件上传api 路由 /api/file/upload 返回一个包含文件存储信息的json数据
func (this *ApiController) FileUpload() {
	info := &ResultData{Error: 1, Title: "失败:", Msg: "上传失败！"}
	f, h, err := this.GetFile("file")
	if err != nil {
		log.Fatal("error: ", err)
	}
	defer f.Close()
	//获取当前年月日
	year, month, _ := tools.EnumerateDate()
	savePath := "file/" + year + "/" + month + "/"
	//创建存储目录
	tools.DirCreate(savePath)
	//重命名文件名称
	tempFileName := tools.StringToMd5(h.Filename, 5)
	suffix := tools.GetFileSuffix(h.Filename)
	saveName := tempFileName + suffix
	// 保存位置
	err = this.SaveToFile("file", savePath+saveName)
	//写入数据库
	if err == nil {
		//写入数据库
		data := &models.Attachment{Name: saveName, Path: savePath + saveName, Created: tools.Int64ToString(time.Now().Unix())}
		id, code := models.FileSave(data)
		if code != nil {
			info = &ResultData{Error: 1, Title: "结果:", Msg: "上传失败！"}
		} else {
			info = &ResultData{Error: 0, Title: "结果:", Msg: "上传成功！", Data: models.FileInfo(id)}
		}
	}
	this.Data["json"] = info
	this.ServeJSON()
}

//文件列表api 路由 /api/file/list
func (this *ApiController) FileList() {
	this.Data["json"] = &JsonData{Code: 0, Count: 100, Msg: "", Data: models.GetFileJson()}
	this.ServeJSON()
}

// 文件删除 路由 /api/file/delete
func (this *ApiController) FileDelete() {
	info := &ResultData{}
	id, _ := strconv.Atoi(this.GetString("id"))
	//数据库文件删除
	filePath, err := models.FileDelete(id)
	if err != nil {
		info = &ResultData{Error: 1, Title: "失败:", Msg: "数据库操作出错！"}
	} else {
		info = &ResultData{Error: 0, Title: "成功:", Msg: "删除成功！"}
	}
	//本地文件删除
	tools.FileRemove(filePath)
	this.Data["json"] = info
	this.ServeJSON()
}

// 七牛云文件上传接口 路由 /api/upload/qiniu
func (this *ApiController) QiniuUpload() {
	f, h, err := this.GetFile("attachment")
	if err != nil {
		log.Fatal("error: ", err)
	}
	defer f.Close()
	fileName := this.GetString("customName") //自定义文件名
	saveName := ""                           //文件存储名
	if fileName == "" {
		saveName = h.Filename
	} else {
		fileSuffix := path.Ext(h.Filename) //得到文件后缀
		saveName = fileName + fileSuffix
	}
	filePath := "file/" + saveName
	this.SaveToFile("attachment", filePath)                           //保存文件到本地
	res := tools.QiniuApi(filePath, saveName, models.SiteConfigMap()) //上传到七牛云
	var data *ResultData
	if res == true {
		data = &ResultData{Error: 1, Title: "结果:", Msg: "上传成功！"}
	} else {
		data = &ResultData{Error: 0, Title: "结果:", Msg: "认证失败！请确保配置信息正确"}
	}
	os.Remove(filePath) //移除本地文件
	this.Data["json"] = data
	this.ServeJSON()
}

// 七牛云文件列表接口 路由 /api/file/qiniu/list
func (this *ApiController) QiniuList() {
	data := models.SiteConfigMap()
	data["Host"] = "api.qiniu.com"
	data["Parameter"] = "/v6/domain/list?tbl=" + data["Bucket"]
	data["Url"] = "http://" + data["Host"] + data["Parameter"]
	Bucket := tools.GetBucketData(data)
	r, _ := regexp.Compile("\"([^\"]*)\"")
	match := r.FindString(string(Bucket))
	match = strings.Replace(match, "\"", "", -1)
	data["Host"] = "rsf.qbox.me"
	data["Parameter"] = "/list?bucket=" + data["Bucket"]
	data["Url"] = "http://" + data["Host"] + data["Parameter"]
	body := tools.GetBucketData(data)
	var res Response
	err := json.Unmarshal([]byte(body), &res)
	if err != nil {
		log.Printf("err was %v", err)
	}
	this.Data["json"] = JsonData{Msg: match, Data: res.Items}
	this.ServeJSON()
}

// 七牛云文件删除 路由 /api/file/qiniu/delete
func (this *ApiController) QiniuDeleteFile() {
	code := this.GetString("code")
	code = base64.StdEncoding.EncodeToString([]byte(code))
	data := models.SiteConfigMap()
	data["Host"] = "rs.qiniu.com"
	data["Parameter"] = "/delete/" + code
	data["Url"] = "http://" + data["Host"] + data["Parameter"]
	this.Data["json"] = JsonData{Data: tools.DeleteFile(data)}
	this.ServeJSON()
}
