/**
 * 数据日志，完成数据上报过程中的日志记录功能
 */
package pipe

import (
	"github.com/donnie4w/go-logger/logger"
)

var LogInsertSQL string = "insert into tb_job_log(logid,ip,jobcode,jobname,rsynctype,dataname,syscode,sysname,platform,Producer,logType,beginTime,flag,endTime,succNum,failNum,errmsg)values(uuid(),'1270.0.0.1','0',?,'04',?,?,?,?,?,'0',?,?,?,'0','0',?)"

/**
 *记录数据日志
 *------ flag 1 成功 0 失败
 *------jobname dataname syscode sysname platform producer beginTime flag endTime errmsg
 */
func DataLog(dbIni *DBini, xmlInfo *XmlInfo, begtime string, err error) {
	stmt, err := dbIni.T_Db.db.Prepare(LogInsertSQL)
	defer stmt.Close()
	if err != nil {
		logger.Error(err)
		return
	}
	scans := make([]interface{}, 10)
	var flag string = "1"
	var msg string
	if err != nil {
		flag = "0"
		msg = string(err.Error())
	}
	scans[0] = xmlInfo.Remark
	scans[1] = xmlInfo.Remark
	scans[2] = dbIni.SysId
	scans[3] = dbIni.SysName
	scans[4] = dbIni.Platform
	scans[5] = dbIni.Company
	scans[6] = begtime
	scans[7] = flag
	scans[8] = NowTime()
	scans[9] = msg
	stmt.Exec(scans...)
}
