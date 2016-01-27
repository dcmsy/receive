/*
* 可变业务数据数据管道，完成读取配置文件，从源库到目的库数据传输功能
 */
package pipe

import (
	"github.com/donnie4w/go-logger/logger"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

/*执行数据同步读取直接写入 single data 单次执行可变的业务数据 */
func (xmlInfo *XmlInfo) RunSingleDataPipe(dbIni *DBini) {
	logger.Info("All run begin >>>>>>>>>>>>>>>>")
	logger.Info("dbIni.F_DbType:", dbIni.F_DbType, "dbIni.T_DbType:", dbIni.T_DbType)
	if !StartFlag {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			logger.Error("查询数据库失败", err)
		}
	}()
	f_pwd, _ := Base64Dec(dbIni.F_Password)
	t_pwd, _ := Base64Dec(dbIni.T_Password)
	dbIni.F_Db.Db_open(dbIni.F_DbType, dbIni.F_User, f_pwd, dbIni.F_Ip, dbIni.F_Port, dbIni.F_Dbname)
	dbIni.T_Db.Db_open(dbIni.T_DbType, dbIni.T_User, t_pwd, dbIni.T_Ip, dbIni.T_Port, dbIni.T_Dbname)
	//关闭数据库
	defer dbIni.F_Db.db.Close()
	defer dbIni.T_Db.db.Close()

	if dbIni.F_Db.db == nil || dbIni.T_Db.db == nil {
		logger.Error("无法链接两方数据库")
		panic("无法链接两方数据库")
		return
	}
	DataSinglePipe(dbIni, xmlInfo)
	Updater(dbIni, xmlInfo)
	ExecProc(dbIni, xmlInfo)
	logger.Info("All run end >>>>>>>>>>>>>>>>")
	time.Sleep(5 * time.Second)
}

/**
 * 边读边写边更新 节省服务器内存开销
 * 中间库的rptdb数据库逻辑 根据查询sql获取i、d状态的数据，逐行插入到rptdb中，插入时判断如果状态为i则insert，状态为d则执行delete语句。
 * 最后更新标记位，该数据记录已读取。
 */
func DataSinglePipe(dbIni *DBini, xmlInfo *XmlInfo) {
	logger.Info("***************************")
	logger.Info("Reader begin >>>>>>>>>>>>>>>>")
	var MatchOrgMap = make(map[string]string)
	var NoneOrgMap = make(map[string]string)
	var begtime string = NowTime()
	var logErr error

	defer func() {
		if err := recover(); err != nil {
			logger.Error("查询数据库失败", err)
		}
	}()

	query, err := dbIni.F_Db.db.Query(xmlInfo.DataSQL) //查询数据库
	if err != nil {
		logger.Error("查询数据库失败", err.Error())
		panic(err.Error())
		return
	}
	defer query.Close()
	stmt, err := dbIni.T_Db.db.Prepare(xmlInfo.InsertSQL)
	defer stmt.Close()
	if err != nil {
		logger.Error(err)
		return
	}

	deleteByIdStmt, err := dbIni.T_Db.db.Prepare(xmlInfo.DeleteByIDSQL)
	defer deleteByIdStmt.Close()
	if err != nil {
		logger.Error(err)
		return
	}
	cols, _ := query.Columns()
	colsLength := len(cols)
	values := make([][]byte, colsLength)
	scans := make([]interface{}, colsLength)
	for i := range values {
		scans[i] = &values[i]
	}
	results := make(map[int]map[string]string)
	idsMap := make(map[int]string)
	i := 0
	for query.Next() { //循环
		if err := query.Scan(scans...); err != nil {
			logger.Error(err)
			logErr = err
			return
		}
		row := make(map[string]string) //每行数据
		for k, v := range values {
			key := cols[k]
			row[key] = string(v)
		}
		results[i] = row //装入结果集中
		wscans := make([]interface{}, colsLength+2)
		deleteByIdScans := make([]interface{}, 1)
		for j := range cols {
			wscans[j] = row[cols[j]]
		}
		//执行组织机构匹配
		if len(xmlInfo.MatchOrgName) > 0 {
			deptname := row[xmlInfo.MatchOrgName]

			//就近原则 当前批次查找
			v, ok := MatchOrgMap[deptname]
			if ok {
				wscans[colsLength] = v
				wscans[colsLength+1] = OrgNameMap[v]
			} else {
				stdorgid, stdorgname := MatchOrg(deptname)
				if strings.EqualFold(stdorgid, noneid) { //未匹配上
					NoneOrgMap[deptname] = noneid
					wscans[colsLength] = noneid
					wscans[colsLength+1] = nonename
				} else { //匹配上了
					MatchOrgMap[deptname] = stdorgid
					wscans[colsLength] = stdorgid
					wscans[colsLength+1] = stdorgname
				}
			}
		} else {
			wscans[colsLength] = ""
			wscans[colsLength+1] = ""
		}

		//判断是否为d 只删除不插入新数据
		deleteByIdScans[0] = row[cols[0]]
		deleteByIdStmt.Exec(deleteByIdScans...)
		if !strings.EqualFold(row["oper_flag"], "d") {
			r, err := stmt.Exec(wscans...)
			if err != nil {
				logErr = err
				logger.Error("#DataSinglePipe装入结果集中,执行insert:", r, err)
			}
		}

		idsMap[i] = row[cols[0]]
		i++
	}

	//insert stdorgmap
	insertStdOrgMap(dbIni, NoneOrgMap)
	//判断是否有数据
	if i == 0 {
		xmlInfo.NotExistData = true
	} else {
		xmlInfo.NotExistData = false
		//insert data log
		DataLog(dbIni, xmlInfo, begtime, logErr)
	}
	mapLength := len(idsMap)
	if mapLength > 0 {
		arrIds := make([]string, mapLength)
		for k, v := range idsMap {
			arrIds[k] = v
		}
		xmlInfo.Ids = arrIds
	}
	idsMap = nil
	logger.Info("Reader end >>>>>>>>>>>>>>>>")
}
