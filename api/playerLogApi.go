package api

import (
	"dst-dashboard/entity"
	"dst-dashboard/vo"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func PlayerLogQueryPage(ctx *gin.Context) {

	//获取查询参数
	name := ctx.Query("name")
	kuId := ctx.Query("kuId")
	steamId := ctx.Query("steamId")

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 10 {
		size = 10
	}

	db := entity.DB

	if name, isExist := ctx.GetQuery("name"); isExist {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if kuId, isExist := ctx.GetQuery("kuId"); isExist {
		db = db.Where("ku_id LIKE ?", "%"+kuId+"%")
	}
	if steamId, isExist := ctx.GetQuery("steamId"); isExist {
		db = db.Where("steamId LIKE ?", "%"+steamId+"%")
	}

	db = db.Order("created_at desc").Limit(size).Offset((page - 1) * size)

	playerLogs := make([]entity.PlayerLog, 0)

	if err := db.Find(&playerLogs).Error; err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("name:", name, "kuId", kuId, "steamId", steamId)

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: playerLogs,
	})

}
