/*
*完成系统注册
 */
package pipe

import (
	"github.com/donnie4w/go-logger/logger"
	"io/ioutil"
	"net/http"
)

var RegisterInsertSQL string = "insert into tb_register (sysCode,sysname,shortname,authCode,platform,producer,isencrypt,encryptmodel,systype,creator,state,addTime,operator,lasttime,valid,impdate)values(?,?,?,uuid(),?,?,'0','AES',?,'go','1',now(),'go',now(),'1',now())"
var RegisterQuerySQL string = "select 1 from tb_register where sysname=?"

/*
 *注册组织机构信息
 */
func (sys *SysIni) RegisterSys() {
	if !StartFlag {
		return
	}
	dbIni := &sys.dbini
	t_pwd, _ := Base64Dec(dbIni.T_Password)
	dbIni.T_Db.Db_open(dbIni.T_DbType, dbIni.T_User, t_pwd, dbIni.T_Ip, dbIni.T_Port, dbIni.T_Dbname)
	defer dbIni.T_Db.db.Close()
	query, err := dbIni.T_Db.db.Query(RegisterQuerySQL, dbIni.SysName) //查询数据库
	if err != nil {
		logger.Error("查询数据库失败", err.Error())
		return
	}
	defer query.Close()
	stmt, err := dbIni.T_Db.db.Prepare(RegisterInsertSQL)
	defer stmt.Close()
	if err != nil {
		logger.Error(err)
		return
	}
	if !query.Next() {
		scans := make([]interface{}, 6)
		scans[0] = dbIni.SysId
		scans[1] = dbIni.SysName
		scans[2] = dbIni.SysName
		scans[3] = dbIni.Platform
		scans[4] = dbIni.Company
		scans[5] = dbIni.SysType
		stmt.Exec(scans...)
	}
	//调用http服务进行系统注册
	var configInfo ConfigInfo = InitConfigInfo()
	var rstUrl = "http://" + configInfo.core_ip + ":" + configInfo.core_port + configInfo.register_url
	if response, err := http.Get(rstUrl); err != nil {
		logger.Error(" err = ", err)
	} else {
		defer response.Body.Close()
		if _, err := ioutil.ReadAll(response.Body); err != nil {
			logger.Error(" err = ", err)
		}
	}
}
