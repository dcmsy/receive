package pipe

import (
	"encoding/json"
	"github.com/astaxie/beego"
	goconfig "github.com/dcmsy/receive/goconfig"
	"github.com/donnie4w/go-logger/logger"
	"strings"
)

//配置信息
type IndexInfo struct {
	beego.Controller
	Singletimer  string
	Repeatetimer string
	Transferes   string
	Nosql        string

	Core_ip        string
	Core_port      string
	Rpt_ip         string
	Rpt_port       string
	Orgversion_url string
	Allorg_url     string
	Register_url   string

	Systype    string
	F_ip       string
	F_port     string
	F_username string
	F_password string

	T_ip       string
	T_port     string
	T_username string
	T_password string
	Sysid      string
	Sysname    string
}

// Ret entity
type Ret struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

//index
func (this *IndexInfo) Index() {
	this.TplNames = "conf/index.html"
}

//帮助
func (this *IndexInfo) Help() {
	this.TplNames = "conf/help.html"
}

//列表
func (this *IndexInfo) List() {
	config := new(IndexInfo)
	configMap := ReadIniFile("config/config.ini", "config")
	GetConfigFromMap(config, configMap)
	l := len(SysMap)
	list := make([]*IndexInfo, l, l)
	list2 := make([]*IndexInfo, l, l)
	for i := range SysArr {
		itemes := strings.Split(SysArr[i], "|")
		info := new(IndexInfo)
		if len(itemes) == 4 {
			info.Systype = itemes[0]
			info.Sysname = itemes[1]
		}
		if i < 7 {
			list[i] = info
		} else {
			list2[i] = info
		}
	}

	this.Data["list"] = list
	this.Data["list2"] = list2
	this.Data["config"] = config
	this.TplNames = "conf/conf.html"
}

//编辑
func (this *IndexInfo) Edit() {
	var (
		config = new(IndexInfo)
		err    error
		ret    Ret
		flag   bool
	)
	if err = json.Unmarshal([]byte(this.GetString("json")), &config); err == nil {
		err = config.SaveAllConf()
	}
	if err != nil {
		logger.Error(err)
		ret.Code, ret.Msg = 1, err.Error()
		flag = false
		logger.Info("Edit", err.Error())
	} else {
		ret.Code = 0
		flag = true
	}
	this.Data["json"] = &ret
	this.ServeJson(flag)
}

//测试连接
func (this *IndexInfo) Test() {
	var (
		config     = new(IndexInfo)
		err        error
		ret        Ret
		db                = new(My_db)
		dbtype     string = "mysql"
		dbusername string
		dbpassowrd string
		dbhostsip  string
		dbport     string
		dbname     string
	)
	if err = json.Unmarshal([]byte(this.GetString("json")), &config); err == nil {
		if strings.EqualFold(config.Systype, "f") {
			dbusername = config.F_username
			dbpassowrd, _ = Base64Dec(config.F_password)
			dbhostsip = config.F_ip
			dbport = config.F_port
			dbname = "mid"
		} else {
			dbusername = config.T_username
			dbpassowrd, _ = Base64Dec(config.T_password)
			dbhostsip = config.T_ip
			dbport = config.T_port
			dbname = "rptdb"
		}
		db.Db_open(dbtype, dbusername, dbpassowrd, dbhostsip, dbport, dbname)
	}
	if err != nil {
		logger.Error(err)
		ret.Code, ret.Msg = 1, err.Error()
	} else {
		if db.db != nil {
			defer db.db.Close()
			query, err := db.db.Query("select 1")
			if err != nil {
				ret.Code = 1
			} else {
				if query.Next() {
					ret.Code = 0
				} else {
					ret.Code = 1
				}
				defer query.Close()
			}
		} else {
			ret.Code = 1
		}
	}
	this.Data["json"] = &ret
	this.ServeJson()
}

//保存所有配置文件
func (this *IndexInfo) SaveAllConf() (err error) {
	var (
		cfgPath string = "config/config.ini"
		retErr  error
	)
	xmlInies := GetXmlIniNames()
	cfg, _ := goconfig.LoadConfigFile(cfgPath)
	IndexInfoToCfg(this, cfg)
	retErr = goconfig.SaveConfig(cfg, cfgPath)
	for _, v := range xmlInies {
		xmlIniCfg, _ := goconfig.LoadConfigFile(v)
		IndexInfoToIni(this, xmlIniCfg)
		retErr = goconfig.SaveConfig(xmlIniCfg, v)
	}
	return retErr
}

//indexinfo to cfg
func IndexInfoToCfg(this *IndexInfo, cfg *goconfig.Config) {

	repeatetimer, _ := cfg.Get("config::repeatetimer")
	singletimer, _ := cfg.Get("config::singletimer")
	register_url, _ := cfg.Get("config::register_url")
	orgversion_url, _ := cfg.Get("config::orgversion_url")
	allorg_url, _ := cfg.Get("config::allorg_url")
	sites, _ := cfg.Get("config::sites")

	cfg.Set("config", "repeatetimer", repeatetimer)
	cfg.Set("config", "singletimer", singletimer)
	cfg.Set("config", "sites", sites)

	cfg.Set("config", "register_url", register_url)
	cfg.Set("config", "orgversion_url", orgversion_url)
	cfg.Set("config", "allorg_url", allorg_url)

	cfg.Set("config", "core_ip", this.Core_ip)
	cfg.Set("config", "core_port", this.Core_port)
	cfg.Set("config", "rpt_ip", this.Rpt_ip)
	cfg.Set("config", "rpt_port", this.Rpt_port)

	cfg.Set("config", "f_ip", this.F_ip)
	cfg.Set("config", "f_port", this.F_port)
	cfg.Set("config", "f_username", this.F_username)
	cfg.Set("config", "f_password", this.F_password)

	cfg.Set("config", "t_ip", this.T_ip)
	cfg.Set("config", "t_port", this.T_port)
	cfg.Set("config", "t_username", this.T_username)
	cfg.Set("config", "t_password", this.T_password)

	cfg.Set("config", "transferes", this.Transferes)
	cfg.Set("config", "nosql", this.Nosql)
}

//indexinfo to ini
func IndexInfoToIni(this *IndexInfo, cfg *goconfig.Config) {
	f_dburl_arr := []string{"jdbc:mysql://", this.F_ip, ":", this.F_port, "/mid?useUnicode=true&characterEncoding=utf-8&useOldAliasMetadataBehavior=true"}
	var f_dburl = strings.Join(f_dburl_arr, "")

	t_dburl_arr := []string{"jdbc:mysql://", this.T_ip, ":", this.T_port, "/rptdb?useUnicode=true&characterEncoding=utf-8&useOldAliasMetadataBehavior=true"}
	var t_dburl = strings.Join(t_dburl_arr, "")
	f_dbname, _ := cfg.Get("dbconfig::f_dbname")
	t_dbname, _ := cfg.Get("dbconfig::t_dbname")

	f_dbtype, _ := cfg.Get("dbconfig::f_dbtype")
	f_driver, _ := cfg.Get("dbconfig::f_driver")
	t_dbtype, _ := cfg.Get("dbconfig::t_dbtype")
	t_driver, _ := cfg.Get("dbconfig::t_driver")
	cfg.Set("dbconfig", "f_dbtype", f_dbtype)
	cfg.Set("dbconfig", "f_driver", f_driver)
	cfg.Set("dbconfig", "f_dburl", f_dburl)
	cfg.Set("dbconfig", "f_ip", this.F_ip)
	cfg.Set("dbconfig", "f_port", this.F_port)
	cfg.Set("dbconfig", "f_dbname", f_dbname)
	cfg.Set("dbconfig", "f_user", this.F_username)
	cfg.Set("dbconfig", "f_password", this.F_password)

	cfg.Set("dbconfig", "t_dbtype", t_dbtype)
	cfg.Set("dbconfig", "t_driver", t_driver)
	cfg.Set("dbconfig", "t_dburl", t_dburl)
	cfg.Set("dbconfig", "t_ip", this.T_ip)
	cfg.Set("dbconfig", "t_port", this.T_port)
	cfg.Set("dbconfig", "t_dbname", t_dbname)
	cfg.Set("dbconfig", "t_user", this.T_username)
	cfg.Set("dbconfig", "t_password", this.T_password)

	sysid, _ := cfg.Get("dbconfig::sysid")
	systype, _ := cfg.Get("dbconfig::systype")
	sysname, _ := cfg.Get("dbconfig::sysname")
	company, _ := cfg.Get("dbconfig::company")
	platform, _ := cfg.Get("dbconfig::platform")
	logger.Info("platform===========================", platform, sysname)
	cfg.Set("dbconfig", "sysid", sysid)
	cfg.Set("dbconfig", "systype", systype)
	cfg.Set("dbconfig", "sysname", sysname)
	cfg.Set("dbconfig", "company", company)
	cfg.Set("dbconfig", "platform", platform)

}

//map to struct
func GetConfigFromMap(config *IndexInfo, cfgMap map[string]string) {
	config.Singletimer = cfgMap["singletimer"]
	config.Repeatetimer = cfgMap["repeatetimer"]
	config.Transferes = cfgMap["transferes"]
	config.Nosql = cfgMap["nosql"]

	config.Core_ip = cfgMap["core_ip"]
	config.Core_port = cfgMap["core_port"]
	config.Rpt_ip = cfgMap["rpt_ip"]
	config.Rpt_port = cfgMap["rpt_port"]

	config.Orgversion_url = cfgMap["orgversion_url"]
	config.Allorg_url = cfgMap["allorg_url"]
	config.Register_url = cfgMap["register_url"]

	config.F_ip = cfgMap["f_ip"]
	config.F_port = cfgMap["f_port"]
	config.F_username = cfgMap["f_username"]
	config.F_password, _ = Base64Dec(cfgMap["f_password"])

	config.T_ip = cfgMap["t_ip"]
	config.T_port = cfgMap["t_port"]
	config.T_username = cfgMap["t_username"]
	config.T_password, _ = Base64Dec(cfgMap["t_password"])
	if len(config.Transferes) > 0 && len(config.Nosql) > 0 {
		config.Systype = config.Transferes + "," + config.Nosql
	} else if len(config.Transferes) > 0 {
		config.Systype = config.Transferes
	} else if len(config.Nosql) > 0 {
		config.Systype = config.Nosql
	} else {
		config.Systype = ""
	}
}
