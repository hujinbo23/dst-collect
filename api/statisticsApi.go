package api

import (
	"dst-dashboard/entity"
	"dst-dashboard/utils"
	"dst-dashboard/vo"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UserStatistics struct {
	Count int       `json:"y"`
	Date  time.Time `json:"x"`
}

type TopStatistics struct {
	Id         int    `json:"id"`
	Count      int    `json:"count"`
	Name       string `json:"name"`
	KuId       string `json:"kuId"`
	SteamId    string `json:"steamId"`
	Role       string `json:"role"`
	ActionDesc string `json:"actionDesc"`
	CreatedAt  string `json:"createdAt"`
}

type RoleRateStatistics struct {
	Role  string `json:"role"`
	Count int    `json:"count"`
}

func find_stamp(stamp int64, data []UserStatistics) *UserStatistics {
	for _, d := range data {
		unix := utils.Bod(d.Date).UnixMilli()
		if unix == stamp {
			return &d
		}
	}
	return nil
}

func CountActiveUser(ctx *gin.Context) {

	unit := ctx.Query("unit")
	startDate := startDate(ctx)
	endDate := endDate(ctx)
	fmt.Println("unit", unit, "startTime", startDate, "endTime", endDate)

	db := entity.DB
	var data1 []UserStatistics
	var data2 []UserStatistics
	var stamps []int64
	//db.Raw("select count(distinct name), day(create_at) from player_logs  where create_at between ? and ? group by month(create_at), day(create_at)", "2023-02-25T16:24:33.2960449+08:00", "2023-02-25T15:59:15.5348647+08:00").Scan(&data)
	if unit == "MONTH" {
		db.Raw("select count(distinct name) as count,created_at as date from player_logs where created_at between ? and ? group by strftime('%Y',created_at),strftime('%m',created_at)", startDate, endDate).Scan(&data1)
		db.Raw("select count(name) as count,created_at as date from player_logs where created_at between ? and ? and action like '[JoinAnnouncement]' group by strftime('%Y',created_at),strftime('%m',created_at)", startDate, endDate).Scan(&data2)
	}
	if unit == "DAY" {
		sql1 := `
		select
			count(name) as count,created_at as date
		from player_logs
		where created_at between ? and ?
		group by strftime('%m',created_at),strftime('%d',created_at)
		`
		sql2 := `
		select
			count(name) as count,created_at as date
		from player_logs
		where created_at between ? and ? and action like '[JoinAnnouncement]'
		group by strftime('%m',created_at),strftime('%d',created_at)
		`
		db.Raw(sql1, startDate, endDate).Scan(&data1)
		db.Raw(sql2, startDate, endDate).Scan(&data2)

		// db.Raw("select count(distinct name) as count,created_at as date from player_logs where action like '[JoinAnnouncement]' group by strftime('%m',created_at),strftime('%d',created_at)").Scan(&data1)
		// db.Raw("select count(name) as count,created_at as date from player_logs where action like '[JoinAnnouncement]' group by strftime('%m',created_at),strftime('%d',created_at)").Scan(&data2)

		stamps = utils.Get_stamp_day(startDate, endDate)
	}

	var axis struct {
		X  []int64 `json:"x"`
		Y1 []int   `json:"y1"`
		Y2 []int   `json:"y2"`
	}
	fmt.Println("data1", data1)
	fmt.Println("data1", data2)
	//填充数据
	// var stamp []int64;
	for _, stamp := range stamps {
		axis.X = append(axis.X, stamp)

		if d := find_stamp(stamp, data1); d != nil {
			axis.Y1 = append(axis.Y1, d.Count)
		} else {
			axis.Y1 = append(axis.Y1, 0)
		}
		if d := find_stamp(stamp, data2); d != nil {
			axis.Y2 = append(axis.Y2, d.Count)
		} else {
			axis.Y2 = append(axis.Y2, 0)
		}
	}

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: axis,
	})
}

func CountLoginUser(ctx *gin.Context) {

	unit := ctx.Query("unit")
	startDate := startDate(ctx)
	endDate := endDate(ctx)
	fmt.Println("unit", unit, "startTime", startDate, "endTime", endDate)

	db := entity.DB
	var data []UserStatistics
	if unit == "MONTH" {
		db.Raw("select count(name) as count,created_at as date from player_logs where created_at between ? and ? group by strftime('%Y',created_at),strftime('%m',created_at)", startDate, endDate).Scan(&data)
	}
	if unit == "DAY" {
		db.Raw("select count(name) as count,created_at as date from player_logs where created_at between ? and ? group by strftime('%m',created_at),strftime('%d',created_at)", startDate, endDate).Scan(&data)
	}

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func TopUserGameTime() {

}

func TopUserActiveTimes(ctx *gin.Context) {
	N := ctx.Query("N")

	// startTime, _ := time.Parse("2006-01-02T15:04:05.000Z", startDate)
	// endTime, _ := time.Parse("2006-01-02T15:04:05.000Z", endDate)

	startDate := startDate(ctx)
	endDate := endDate(ctx)

	fmt.Println("N", N, "startTime", startDate, "endTime", endDate)

	db := entity.DB

	//本天，本周，本月
	var data []TopStatistics
	sql := `
	select 
		max(id) as id,count(distinct name) as count, name, ku_id, steam_id, role, action_desc, created_at
	from player_logs
	where created_at between ? and ?
	group by name order by count(id) DESC limit ?
	`
	db.Raw(sql, startDate, endDate, N).Scan(&data)

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func TopUserLoginimes(ctx *gin.Context) {
	N := ctx.Query("N")
	startDate := startDate(ctx)
	endDate := endDate(ctx)

	fmt.Println("N", N, "startDate", startDate, "endDate", endDate)

	db := entity.DB

	//本天，本周，本月
	var data []TopStatistics
	sql := `
	select 
		max(p.id) as id,count(p.name) as count, p.name, c.ku_id, c.steam_id, role, action_desc, p.created_at
	from player_logs p
	left join connects c on p.name = c.name
	where p.created_at between ? and ? and p.action like '[JoinAnnouncement]' 
	group by p.name order by count(p.id) DESC limit ?
	`
	db.Raw(sql, startDate, endDate, N).Scan(&data)

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func TopDeaths(ctx *gin.Context) {

	N := ctx.Query("N")
	startDate := startDate(ctx)
	endDate := endDate(ctx)

	fmt.Println("N", N, "startDate", startDate, "endDate", endDate)

	db := entity.DB

	//本天，本周，本月
	var data []TopStatistics
	sql := `
	select 
		max(id) as id, count(id) as count, name, ku_id, steam_id, role, action_desc, created_at
	from player_logs
	where created_at between ? and ? and action like '[DeathAnnouncement]'
	group by name order by count(id) DESC limit ?
	`
	db.Raw(sql, startDate, endDate, N).Scan(&data)

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func CountRoleRate(ctx *gin.Context) {
	startDate := startDate(ctx)
	endDate := endDate(ctx)

	db := entity.DB

	//本天，本周，本月
	var data []RoleRateStatistics
	sql := `
	select 
		role as role, count(distinct name) as count
	from player_logs
	where created_at between ? and ?
	group by role
	`
	db.Raw(sql, startDate, endDate).Scan(&data)

	ctx.JSON(http.StatusOK, vo.Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func startDate(ctx *gin.Context) time.Time {
	var date time.Time
	if t, isExist := ctx.GetQuery("startDate"); isExist {
		date, _ = time.Parse("2006-01-02T15:04:05.000Z", t)
	}
	return date
}

func endDate(ctx *gin.Context) time.Time {
	var date time.Time
	if t, isExist := ctx.GetQuery("endDate"); isExist {
		date, _ = time.Parse("2006-01-02T15:04:05.000Z", t)
	}
	return date
}
