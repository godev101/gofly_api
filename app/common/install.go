package common

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gofly/model"
	"gofly/utils"
	"gofly/utils/results"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

/**
* 项目安装
 */
type Install struct {
}

func init() {
	fpath := Install{}
	utils.Register(&fpath, reflect.TypeOf(fpath).PkgPath())
}

// 安装页面
func (api *Install) Index(context *gin.Context) {
	path, err := os.Getwd()
	if err != nil {
		results.Failed(context, "项目路径获取失败", nil)
		return
	}
	//filePath := fmt.Sprintf("%s\\resource\\staticfile\\template\\install.lock", path)
	filePath := filepath.Join(path, "/resource/staticfile/template/install.lock")
	if _, err := os.Stat(filePath); err == nil {
		context.HTML(http.StatusOK, "isinstall.html", gin.H{
			"title": "已经安装页面",
		})
	} else {
		context.HTML(http.StatusOK, "install.html", gin.H{
			"title": "安装页面",
		})
	}

}

// 安装
func (api *Install) Save(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var parameter map[string]interface{}
	_ = json.Unmarshal(body, &parameter)
	path, err := os.Getwd()
	if err != nil {
		results.Failed(c, "项目路径获取失败", nil)
		return
	}
	model.CreateDataBase(parameter["username"], parameter["password"], parameter["hostname"], parameter["hostport"], parameter["database"])
	//2.修改数据库配置
	cferr := upConfFieldData(path, parameter)
	if cferr != nil {
		results.Failed(c, "修改数据库配置失败", nil)
		return
	}
	model.MyInit(2) //初始化数据
	//创建数据库

	//导入书库配置
	//SqlPath := fmt.Sprintf("%v\\resource\\staticfile\\template\\gofly_api.sql", path)
	SqlPath := filepath.Join(path, "/resource/staticfile/template/gofly_api.sql")
	sqls, sqlerr := os.ReadFile(SqlPath)
	if sqlerr != nil {
		results.Failed(c, "数据库文件不存在："+SqlPath, nil)
		return
	}
	sqlArr := strings.Split(string(sqls), ";")
	for _, sql := range sqlArr {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		model.ExecSql(sql)
	}
	//创建安装锁文件
	//filePath := fmt.Sprintf("%s\\resource\\staticfile\\template\\install.lock", path)
	filePath := filepath.Join(path, "/resource/staticfile/template/install.lock")
	os.Create(filePath)
	results.Success(c, "安装成功,去前端刷新试试！", parameter, nil)
}

// 更新配置文件
func upConfFieldData(path string, parameter map[string]interface{}) error {
	//file_path := fmt.Sprintf("%v\\config\\settings.yml", path)
	file_path := filepath.Join(path, "/resource/config.yml")
	f, err := os.Open(file_path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	var result = ""
	var is_hose = false
	for {
		is_hose = false
		a, _, c := buf.ReadLine()
		if c == io.EOF {
			break
		}
		for keys, Val := range parameter {
			if strings.Contains(string(a), keys) {
				is_hose = true
				datestr := strings.ReplaceAll(string(a), string(a), fmt.Sprintf("     %v: %v\n", keys, Val))
				result += datestr
			}
		}
		if !is_hose {
			result += string(a) + "\n"
		}
	}
	fw, err := os.OpenFile(file_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666) //os.O_TRUNC清空文件重新写入，否则原文件内容可能残留
	w := bufio.NewWriter(fw)
	w.WriteString(result)
	if err != nil {
		panic(err)
	}
	w.Flush()
	return nil
}
