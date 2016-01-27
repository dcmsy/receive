package pipe

import (
	"database/sql"
	"gopkg.in/mgo.v2/bson"
	"time"
)

/*db 句柄*/
type My_db struct {
	db *sql.DB
}

/*系统配置信息*/
type SysIni struct {
	//db配置信息
	dbini      DBini
	configInfo ConfigInfo
	//上报xml列表
	xmlInfoes       []XmlInfo
	singleXmlInfoes []XmlInfo
	xmlInfo         XmlInfo
}

type ConfigInfo struct {
	singletimer  string
	repeatetimer string
	transferes   string
	nosql        string

	core_ip        string
	core_port      string
	rpt_ip         string
	rpt_port       string
	orgversion_url string
	allorg_url     string
	register_url   string

	f_ip       string
	f_port     string
	f_username string
	f_password string

	t_ip       string
	t_port     string
	t_username string
	t_password string
	sites      string
}

/*数据库配置信息*/
type DBini struct {
	F_Db     My_db
	T_Db     My_db
	F_DbType string
	T_DbType string

	F_Driver   string
	F_Dburl    string
	F_Ip       string
	F_Port     string
	F_Dbname   string
	F_User     string
	F_Password string

	T_Driver   string
	T_Dburl    string
	T_Ip       string
	T_Port     string
	T_Dbname   string
	T_User     string
	T_Password string

	SysId    string
	SysType  string
	SysName  string
	Company  string
	Platform string
}

/*xml配置信息*/
type XmlInfo struct {
	FileName       string `xml:"filename,attr"`
	Remark         string `xml:"remark,attr"`
	RunModel       string `xml:"runmodel,attr"`
	IsValData      string `xml:"isvaldata,attr"`
	MatchOrgName   string `xml:"matchorgname,attr"`
	Isexecpro      string `xml:"isexecpro,attr"`
	Proname        string `xml:"proname,attr"`
	CheckRowSQL    string `xml:"CheckRowSQL"`
	UpdateOperSQL  string `xml:"UpdateOperSQL"`
	DeleteByIDSQL  string `xml:"DeleteByIDSQL"`
	CheckColumnSQL string `xml:"CheckColumnSQL"`
	AlterSQL       string `xml:"AlterSQL"`
	UpdateFlagSQL  string `xml:"UpdateFlagSQL"`
	DataSQL        string `xml:"DataSQL"`
	InsertSQL      string `xml:"InsertSQL"`
	SuccessSQL     string `xml:"SuccessSQL"`
	Ids            []string
	Datas          map[int]map[string]string
	Columns        []string
	GobalNum       int
	RowNum         int
	AlterFlag      bool
	CheckFlag      bool
	NotExistData   bool
}

//站点
type Site struct {
	Id           bson.ObjectId `bson:"_id"`
	Numattention int32         `bson:"numattention"`
	Numnormal    int32         `bson:"numnormal"`
	Numsuspicion int32         `bson:"numsuspicion"`
	Numwarning   int32         `bson:"numwarning"`
	Organization string        `bson:"organization"`
	Remark       string        `bson:"remark"`
	Sitekind     string        `bson:"sitekind"`
	Sitename     string        `bson:"sitename"`
	Siteowner    string        `bson:"siteowner"`
	Siteruleid   string        `bson:"siteruleid"`
	Siteurl      string        `bson:"siteurl"`
	Siteurlhash  int32         `bson:"siteurlhash"`
	State        int32         `bson:"state"`
	Telphone     string        `bson:"telphone"`
}

//alarmprint
type Alarmprint struct {
	Id          bson.ObjectId `bson:"_id"`
	Title       string        `bson:"title"`
	Url         string        `bson:"url"`
	Urlhash     int32         `bson:"urlhash"`
	Keywords    []string      `bson:"keywords"`
	Siteurlhash int32         `bson:"siteurlhash"`
	Time        time.Time     `bson:"time"`
	State       int32         `bson:"state"`
}

//siteChartTable
type SiteChartTable struct {
	Id           bson.ObjectId `bson:"_id"`
	Siteurlhash  int32         `bson:"siteurlhash"`
	Year         string        `bson:"year"`
	Month        string        `bson:"month"`
	Numwarning   int32         `bson:"numwarning"`
	Numsuspicion int32         `bson:"numsuspicion"`
	Numattention int32         `bson:"numattention"`
}

/*组织机构信息*/
type OrgInfo struct {
	Region      string `json:"region"`
	Remark      string `json:"remark"`
	Ip          string `json:"ip"`
	Node_id     string `json:"node_id"`
	Org_path    string `json:"org_path"`
	Id          string `json:"id"`
	Create_date string `json:"create_date"`
	Name        string `json:"name"`
	Parent_id   string `json:"parent_id"`
}
