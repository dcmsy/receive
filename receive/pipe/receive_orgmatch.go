/*
* 可变业务数据数据管道，完成读取配置文件，从源库到目的库数据传输功能
 */
package pipe

import (
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

/*
 *1.每隔20秒检测组织机构的版本变化，如果发生变化则同步平台组织机构信息
 *2.拉取组织机构并生成合适的map数据
 */
var OrgVersion string = ""
var noneid string = "none"
var nonename string = "待匹配组织机构"

var OrgMap = make(map[string]string)
var OrgNameMap = make(map[string]string)
var StdOrgMap = make(map[string]string)

var OrgFlag bool = false
var Time_interval time.Duration = 30
var StdOrgMapInsertSQL string = "insert into tb_stdorg_map(id,std_deptid,map_deptname,isMap,sub_sysid,sub_sysname,creator,addTime,operator,lastTime,impdate,isupdate,sysnames)values(uuid(),'none',?,'0',?,?,'go',now(),'go',now(),now(),'0',?)"
var StdOrgMapQuerySQL string = "select id,sysnames from tb_stdorg_map where map_deptname=?"
var MapQuerySQL string = "select std_deptid,map_deptname from tb_stdorg_map where isMap='1'"
var UpdateSQL string = "update tb_stdorg_map set sysnames=? where id=?"

/*已匹配组织机构信息*/
func (sys *SysIni) StdOrgSync() {
	logger.Info("StdOrgSync start >>>>>>>>>>>>>>>>")
	defer func() {
		if err := recover(); err != nil {
			logger.Error("查询数据库失败", err)
		}
	}()
	for {
		dbIni := &sys.dbini
		t_pwd, _ := Base64Dec(dbIni.T_Password)
		dbIni.T_Db.Db_open(dbIni.T_DbType, dbIni.T_User, t_pwd, dbIni.T_Ip, dbIni.T_Port, dbIni.T_Dbname)
		defer dbIni.T_Db.db.Close()
		query, err := dbIni.T_Db.db.Query(MapQuerySQL) //查询数据库
		if err != nil {
			logger.Error("查询数据库失败", err.Error())
			panic(err)
			return
		}
		defer query.Close()
		for k, _ := range StdOrgMap {
			delete(StdOrgMap, k)
		}

		cols, _ := query.Columns()
		colsLength := len(cols)
		values := make([][]byte, colsLength)
		scans := make([]interface{}, colsLength)
		for i := range values {
			scans[i] = &values[i]
		}
		for query.Next() { //循环
			if err := query.Scan(scans...); err != nil {
				logger.Error(err)
				return
			}
			row := make(map[string]string) //每行数据
			for k, v := range values {
				key := cols[k]
				row[key] = string(v)
			}
			StdOrgMap[row["map_deptname"]] = row["std_deptid"]
		}
		time.Sleep(Time_interval * time.Second)
	}
}

/*启动组织机构同步服务*/
func (configinfo *ConfigInfo) StartOrgSync() {
	logger.Info(" StartOrgSync is run... ")
	var orgVersionUrl = "http://" + configinfo.core_ip + ":" + configinfo.core_port + configinfo.orgversion_url
	var allOrgUrl = "http://" + configinfo.core_ip + ":" + configinfo.core_port + configinfo.allorg_url
	logger.Info("orgVersionUrl start >>>>>>>>>>>>>>>>", orgVersionUrl)
	logger.Info("allOrgUrl start >>>>>>>>>>>>>>>>", allOrgUrl)
	//发起链接访问，获取返回的json数据
	for {
		var return_orgs []byte
		var return_version []byte
		var tmpOrgVersion string = ""
		if response, err := http.Get(orgVersionUrl); err != nil {
			logger.Error(" err = ", err)
		} else {
			defer response.Body.Close()
			if return_version, err = ioutil.ReadAll(response.Body); err != nil {
				logger.Error(" err = ", err)
			}
			tmpOrgVersion = string(return_version)
		}
		logger.Info(" return_version = ", tmpOrgVersion)
		if len(OrgVersion) == 0 || (len(tmpOrgVersion) > 0 && !strings.EqualFold(tmpOrgVersion, OrgVersion)) {
			OrgVersion = tmpOrgVersion
			//重新加载组织机构
			if response, err := http.Get(allOrgUrl); err != nil {
				logger.Error(" err = ", err)
			} else {
				defer response.Body.Close()
				if return_orgs, err = ioutil.ReadAll(response.Body); err != nil {
					logger.Error(" err = ", err)
				}
			}
			var orgInfos []OrgInfo
			err := json.Unmarshal(return_orgs, &orgInfos)
			if err != nil {
				logger.Error(" err = ", err)
			}
			if len(orgInfos) > 0 {
				for k, _ := range OrgMap {
					delete(OrgMap, k)
				}
				for k, _ := range OrgNameMap {
					delete(OrgNameMap, k)
				}
				for i := range orgInfos {
					org := &orgInfos[i]
					OrgMap[org.Org_path] = org.Id
					OrgMap[org.Name] = org.Id
					OrgNameMap[org.Id] = org.Name
				}
			}
		}
		time.Sleep(Time_interval * time.Second)
	}
}

/*
*组织机构匹配
*返回 id name
*regex=<<`>>`/`|`<`>
*replaceTo=\\\\
 */
func MatchOrg(deptname string) (string, string) {
	var stdorgid string = noneid
	var stdorgname string = nonename
	newstdorgid, ok := StdOrgMap[deptname]
	if ok {
		stdorgid = newstdorgid
		stdorgname = OrgNameMap[newstdorgid]
	} else {
		deptname = strings.Replace(deptname, "|", "\\", -1)
		deptname = strings.Replace(deptname, "<<", "\\", -1)
		deptname = strings.Replace(deptname, ">>", "\\", -1)
		deptname = strings.Replace(deptname, "/", "\\", -1)
		deptname = strings.Replace(deptname, ">", "\\", -1)
		deptname = strings.Replace(deptname, "<", "\\", -1)

		v, ok := OrgMap[deptname]
		if ok {
			stdorgid = v
			stdorgname = OrgNameMap[v]
		}
	}
	return stdorgid, stdorgname
}

/*添加待匹配组织机构*/
func insertStdOrgMap(dbIni *DBini, noneOrgMap map[string]string) {
	stdOrgMapStmt, err := dbIni.T_Db.db.Prepare(StdOrgMapInsertSQL)
	defer stdOrgMapStmt.Close()
	if err != nil {
		logger.Error(err)
		return
	}
	for deptname, _ := range noneOrgMap {
		if len(deptname) == 0 {
			continue
		}
		stdOrgMapQuery, err := dbIni.T_Db.db.Query(StdOrgMapQuerySQL, deptname)
		if err != nil {
			logger.Error("查询数据库失败", err.Error())
			return
		}
		defer stdOrgMapQuery.Close()
		if stdOrgMapQuery.Next() { //如果存在 则判断是否需要更新 sysnames
			cols, _ := stdOrgMapQuery.Columns()
			colsLength := len(cols)
			values := make([][]byte, colsLength)
			scans := make([]interface{}, colsLength)
			for i := range values {
				scans[i] = &values[i]
			}
			if err := stdOrgMapQuery.Scan(scans...); err != nil {
				logger.Error(err)
				return
			}
			row := make(map[string]string) //每行数据
			for k, v := range values {
				key := cols[k]
				row[key] = string(v)
			}
			id := row["id"]
			sysnames := row["sysnames"]
			if len(sysnames) > 0 && len(dbIni.SysName) > 0 && !strings.Contains(sysnames, dbIni.SysName) {
				updateStmt, err := dbIni.T_Db.db.Prepare(UpdateSQL)
				defer updateStmt.Close()
				if err != nil {
					logger.Error(err)
					return
				}
				updateScans := make([]interface{}, 2)
				updateScans[0] = sysnames + "," + dbIni.SysName
				updateScans[1] = id
				updateStmt.Exec(updateScans...)
			}
		} else { //insert
			stdOrgScans := make([]interface{}, 4)
			stdOrgScans[0] = deptname
			stdOrgScans[1] = dbIni.SysId
			stdOrgScans[2] = dbIni.SysName
			stdOrgScans[3] = dbIni.SysName
			stdOrgMapStmt.Exec(stdOrgScans...)
		}
	}
}
