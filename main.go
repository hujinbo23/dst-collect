package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"dst-dashboard/api"
	"dst-dashboard/collect"
	"dst-dashboard/config"
	"dst-dashboard/entity"
	"dst-dashboard/middleware"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
)

var server_chat_log string
var server_log string
var db string = "test.db"
var port string = ":8081"

func arg() {
	for idx, args := range os.Args {
		split := strings.Split(args, "=")
		if len(split) == 2 {
			argName := strings.TrimSpace(split[0])
			if argName == "path" {
				argValue := strings.TrimSpace(split[1])
				value := argValue

				server_chat_log = filepath.Join(value, "Master", "server_chat_log.txt")
				server_log = filepath.Join(value, "Master", "server_log.txt")
			}
			if argName == "db" {
				argValue := strings.TrimSpace(split[1])
				value := argValue
				db = value
			}
			if argName == "port" {
				argValue := strings.TrimSpace(split[1])
				value := argValue
				port = ":" + value
			}
		}
		fmt.Println("参数"+strconv.Itoa(idx)+":", args)
	}
	fmt.Println("server_chat_log", server_chat_log)
	fmt.Println("server_log", server_log)
	fmt.Println("db", db)
}

var configData *config.Config

func InitConfig() {
	yamlFile, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		fmt.Println(err.Error())
	}
	var _config *config.Config
	err = yaml.Unmarshal(yamlFile, &_config)
	if err != nil {
		fmt.Println(err.Error())
	}
	configData = _config
	fmt.Println(_config)
}

func iniiDB() {
	db, err := gorm.Open(sqlite.Open(configData.Db), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}
	entity.DB = db
	entity.DB.AutoMigrate(&entity.Spawn{}, &entity.PlayerLog{}, &entity.Connect{})
}

func main() {
	//arg()
	InitConfig()
	iniiDB()

	go tailf_server_log()
	go tailf_server_chat_log()

	app := gin.Default()
	app.Use(middleware.Recover)

	app.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	player := app.Group("/api/player")
	{
		player.GET("/log", api.PlayerLogQueryPage)
	}

	statistics := app.Group("/api/statistics")
	{
		statistics.GET("/active/user", api.CountActiveUser)
		statistics.GET("/top/death", api.TopDeaths)
		statistics.GET("/top/login", api.TopUserLoginimes)
		statistics.GET("/top/active", api.TopUserActiveTimes)

		statistics.GET("/rate/role", api.CountRoleRate)
	}

	app.Run(":" + configData.Port)

}

func tailf_server_chat_log() {
	//fileName := "C:\\Users\\xm\\Documents\\Klei\\DoNotStarveTogether\\900587905\\Cluster_2\\Master\\server_chat_log.txt"
	fileName := filepath.Join(configData.Path, "Master", "server_chat_log.txt")
	config := tail.Config{
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 是否跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的哪个地方开始读
		MustExist: false,                                // 文件不存在不报错
		Poll:      true,
	}
	tails, err := tail.TailFile(fileName, config)
	if err != nil {
		log.Println("文件监听失败")
	}
	var (
		line *tail.Line
		ok   bool
	)
	for {
		line, ok = <-tails.Lines
		if !ok {
			log.Println("文件监听失败")
		}
		//log.Println(line.Text)
		collect.CollectChatLog(line.Text)
	}
}

func tailf_server_log() {
	//fileName := "C:\\Users\\xm\\Documents\\Klei\\DoNotStarveTogether\\900587905\\Cluster_2\\Master\\server_log.txt"
	fileName := filepath.Join(configData.Path, "Master", "server_log.txt")
	config := tail.Config{
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 是否跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的哪个地方开始读
		MustExist: false,                                // 文件不存在不报错
		Poll:      true,
	}
	tails, err := tail.TailFile(fileName, config)
	if err != nil {
		log.Println("文件监听失败")
	}
	var (
		line *tail.Line
		ok   bool
	)
	var perLine string = ""
	var start string = ""
	first := true
	connection := false
	i := 0
	var connect entity.Connect
	for {
		line, ok = <-tails.Lines
		if !ok {
			log.Println("文件监听失败")
		}

		text := line.Text
		perLine = text

		if first {
			start = text
			//fmt.Println("日志", text)
			first = false
		}

		//解析 时间
		if find := strings.Contains(text, "# Generating"); find {
			fmt.Println("房间结束了", start, perLine)
		}
		if find := strings.Contains(text, "Spawn request"); find {
			collect.CollectSpawnRequestLog(text)
		}

		//New incoming connection
		if find := strings.Contains(text, "New incoming connection"); find {
			connection = true
			connect = entity.Connect{}
		}
		if connection {
			if i > 5 {
				i = 0
				connection = false
			} else {
				//do

				if i == 1 {
					// 解析 ip
					fmt.Println(1, text)
					str := strings.Split(text, " ")
					var ip string
					if strings.Contains(text, "[LAN]") {
						ip = str[5]
					} else {
						ip = str[4]
					}
					connect.Ip = ip
					fmt.Println("ip", ip)
				} else if i == 3 {
					fmt.Println(3, text)
					// 解析 KU 和 用户名
					str := strings.Split(text, " ")
					ku := str[3]
					ku = ku[1 : len(ku)-1]
					name := str[4]
					connect.Name = name
					connect.KuId = ku
					fmt.Println("ku", ku, "name", name)
				} else if i == 4 {
					fmt.Println(4, text)
					// 解析 steamId
					str := strings.Split(text, " ")
					steamId := str[4]
					steamId = steamId[1 : len(steamId)-1]
					fmt.Println("steamId", steamId)

					//记录
					connect.SteamId = steamId
					entity.DB.Create(&connect)
				}
				i++
			}
		}

	}
}
