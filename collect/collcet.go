package collect

import (
	"dst-dashboard/entity"
	"fmt"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func StartCollect() {
	//1. 监控文件
	//2. 获取更新的数据
	//3. 处理加工的数据
	//4. 保存数据库
}

func TailfFile() {

}

func CollectChatLog(text string) {
	//[00:00:55]: [Join Announcement] 猜猜我是谁
	if strings.Contains(text, "[Join Announcement]") {
		parseJoin(text)
	}
	//[00:02:28]: [Leave Announcement] 猜猜我是谁
	if strings.Contains(text, "[Leave Announcement]") {
		parseLeave(text)
	}
	//[00:02:17]: [Death Announcement] 猜猜我是谁 死于： 采摘的红蘑菇。她变成了可怕的鬼魂！
	if strings.Contains(text, "[Death Announcement]") {
		parseDeath(text)
	}
	//[00:02:37]: [Resurrect Announcement] 猜猜我是谁 复活自： TMIP 控制台.
	if strings.Contains(text, "[Resurrect Announcement]") {
		parseResurrect(text)
	}
	//[00:03:16]: [Say] (KU_Mt-zrX8K) 猜猜我是谁: 你好啊
	if strings.Contains(text, "[Say]") {
		parseSay(text)
	}
}

func parseSay(text string) {
	fmt.Println(text)

	arr := strings.Split(text, " ")
	temp := strings.Replace(arr[0], " ", "", -1)
	time := temp[:len(temp)-1]
	action := arr[1]
	kuId := arr[2]
	kuId = kuId[1 : len(kuId)-1]
	name := arr[3]
	name = name[:len(name)-1]
	rest := ""
	for i := 4; i <= len(arr)-1; i++ {
		rest += arr[i] + " "
	}
	actionDesc := rest

	spawn := getSpawnRole(name)
	connect := getConnectInfo(name)

	playerLog := entity.PlayerLog{
		Name:       name,
		Role:       spawn.Role,
		Action:     action,
		ActionDesc: actionDesc,
		Time:       time,
		Ip:         connect.Ip,
		KuId:       kuId,
		SteamId:    connect.SteamId,
	}
	//fmt.Println("time", time, "action:", action, "name:", "kuId:", kuId, name, "op:", actionDesc)
	//获取最近的一条spwan记录和newComing
	//playerLog := entity.PlayerLog{Name: name, Role: spawn.Role, KuId: kuId, Action: action, ActionDesc: actionDesc, Time: time}

	entity.DB.Create(&playerLog)

}

func parseResurrect(text string) {
	parseDeath(text)
}

func parseDeath(text string) {
	fmt.Println(text)
	arr := strings.Split(text, " ")

	temp := strings.Replace(arr[0], " ", "", -1)
	time := temp[:len(temp)-1]
	action := arr[1] + arr[2]
	name := strings.Replace(arr[3], "\n", "", -1)

	rest := ""
	for i := 4; i <= len(arr)-1; i++ {
		rest += arr[i] + " "
	}
	actionDesc := rest

	//获取最近的一条spwan记录和newComing
	spawn := getSpawnRole(name)
	connect := getConnectInfo(name)
	fmt.Println(connect)

	playerLog := entity.PlayerLog{
		Name:       name,
		Role:       spawn.Role,
		Action:     action,
		ActionDesc: actionDesc,
		Time:       time,
		Ip:         connect.Ip,
		KuId:       connect.KuId,
		SteamId:    connect.SteamId,
	}

	entity.DB.Create(&playerLog)

}

func parseLeave(text string) {
	parseJoin(text)
}

func parseJoin(text string) {
	fmt.Println(text)
	arr := strings.Split(text, " ")
	temp := strings.Replace(arr[0], " ", "", -1)
	time := temp[:len(temp)-1]
	action := arr[1] + arr[2]
	name := arr[3]

	spawn := getSpawnRole(name)
	connect := getConnectInfo(name)

	playerLog := entity.PlayerLog{
		Name:    name,
		Role:    spawn.Role,
		Action:  action,
		Time:    time,
		Ip:      connect.Ip,
		KuId:    connect.KuId,
		SteamId: connect.SteamId,
	}
	//获取最近的一条spwan记录和newComing
	//playerLog := entity.PlayerLog{Name: name, Role: spawn.Role, Action: action, ActionDesc: "", Time: time}
	entity.DB.Create(&playerLog)
}

func CollectSpawnRequestLog(text string) {
	// Spawn request: wurt from 猜猜我是谁
	arr := strings.Split(text, " ")
	temp := strings.Replace(arr[0], " ", "", -1)
	time := temp[:len(temp)-1]
	role := strings.Replace(arr[3], " ", "", -1)
	name := strings.Replace(arr[5], "\n", "", -1)

	spawn := entity.Spawn{Name: name, Role: role, Time: time}
	entity.DB.Create(&spawn)

}

func getSpawnRole(name string) *entity.Spawn {
	spawn := new(entity.Spawn)
	entity.DB.Where("name LIKE ?", "%"+name+"%").Last(spawn)
	return spawn
}

func getConnectInfo(name string) *entity.Connect {
	connect := new(entity.Connect)
	entity.DB.Where("name LIKE ?", "%"+name+"%").Last(connect)
	return connect
}

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

func ConvertByte2String(byte []byte, charset Charset) string {

	var str string
	switch charset {
	case GB18030:
		decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}
