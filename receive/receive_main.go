package main

import (
	"github.com/donnie4w/go-logger/logger"
	"runtime"
	"time"

	"flag"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	pipe "github.com/dcmsy/receive/pipe"
	"os"
	"path/filepath"
)

/*service name*/
const (
	cSvcName = "rpt_receive"
)

/*beego初始化*/
func init() {
	var config_file string
	flag.StringVar(&config_file, "conf", "", "the path of the config file")
	flag.Parse()
	if config_file != "" {
		beego.AppConfigPath, _ = filepath.Abs(config_file)
		beego.ParseConfig()
	} else {
		if config_file = os.Getenv("BEEGO_APP_CONFIG_FILE"); config_file != "" {
			beego.AppConfigPath, _ = filepath.Abs(config_file)
			beego.ParseConfig()
		}
	}
}

/*主入口*/
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//指定是否控制台打印，默认为true
	logger.SetConsole(true)

	//指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高
	logger.SetLevel(logger.FATAL)
	logger.SetRollingFile("./log", "receivelog.log", 100, 50, logger.MB)
	if beego.AppConfig.String("runmode") == "dev" {
		orm.Debug = true
	}

	//配置文件初始化
	err := pipe.InitFile()
	if err != nil {
		logger.Error("InitFile():", err)
	}

	//注册运行服务
	ServiceHandle(cSvcName, cSvcName, cSvcName, func() {

		//初始化运行参数
		pipe.InitConfigMap()
		//运行参数管理UI服务
		beego.Router("/", &pipe.IndexInfo{}, "*:Index")
		beego.Router("/index/list", &pipe.IndexInfo{}, "*:List")
		beego.Router("/index/edit", &pipe.IndexInfo{}, "*:Edit")
		beego.Router("/index/test", &pipe.IndexInfo{}, "*:Test")
		beego.Router("/index/help", &pipe.IndexInfo{}, "*:Help")

		//获取平台组织机构
		var configInfo pipe.ConfigInfo = pipe.InitConfigInfo()
		go configInfo.StartOrgSync()
		time.Sleep(5 * time.Second)
		logger.Info(" 完成获取基础平台组织机构..... ")

		//启动数据同步
		go pipe.StartAllSys()

		//启动参数管理UI 服务
		beego.Run()

		logger.Info("DaemonName=", cSvcName, "Daemon started.")
		for {
			time.Sleep(time.Hour)
		}
	}, func() {
		Stop()
		logger.Info("DaemonName=", cSvcName, "Daemon stoped.")
	})
}

/*停止数据传输*/
func Stop() {
	pipe.StartFlag = false
}
