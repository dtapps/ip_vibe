package handlers

import (
	"net/http"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/initialize"

	"github.com/gin-gonic/gin"
)

// 获取全部任务
func GetTasks(c *gin.Context) {
	// 查询全部任务
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	successResponse(c, jobHelper.GetJobsData(5))
}

// 获取任务
func GetTask(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询全部任务
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	jobs := jobHelper.GetJobsData(15)
	for _, job := range jobs {
		if job.ID == id {
			successResponse(c, job)
			return
		}
	}
	errorResponse(c, http.StatusNotFound, "任务不存在")
}

// 运行任务
func RunTask(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询全部任务
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 运行任务
	err := jobHelper.RunJob(id)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, map[string]string{"message": "任务运行成功"})
}

// 删除任务
func DeleteTask(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询全部任务
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 删除任务
	err := jobHelper.DeleteJob(id)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, map[string]string{"message": "任务删除成功"})
}

// 停止调度器
func StopTask(c *gin.Context) {
	initialize.StopGoCron(c)
	successResponse(c, map[string]string{"message": "调度器停止成功"})
}

// 启动调度器
func StartTask(c *gin.Context) {
	initialize.StartGoCron(c)
	successResponse(c, map[string]string{"message": "调度器启动成功"})
}
