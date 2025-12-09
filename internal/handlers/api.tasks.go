package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/backend"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 获取全部任务
//
//	@Summary	获取全部任务
//	@Tags		任务
//	@Accept application/json
//	@Produce application/json
//	@Success	200	{object}	APIResponse{Data=[]gocron_helper.JobData}
//	@Failure	404	{object}	APIResponse
//	@Router		/tasks [get]
func GetTasks(c *gin.Context) {
	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 获取全部任务
	jobs := jobHelper.GetJobsData(5)
	successResponse(c, jobs)
}

// 获取任务
func GetTask(c *gin.Context) {
	// 请求参数
	var requestParams types.GetTaskRequest
	requestParams.ID = c.Param("id")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 获取全部任务
	jobs := jobHelper.GetJobsData(15)
	for _, job := range jobs {
		if job.ID == requestParams.ID {
			successResponse(c, job)
			return
		}
	}
	errorResponse(c, http.StatusNotFound, "任务不存在")
}

// 运行任务
func RunTask(c *gin.Context) {
	// 请求参数
	var requestParams types.RunTaskRequest
	requestParams.ID = c.Param("id")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 运行任务
	err := jobHelper.RunJob(requestParams.ID)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("运行任务: %s - %s\n", requestParams.ID, err.Error())
		}
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, map[string]string{"message": "任务运行成功"})
}

// 删除任务
func DeleteTask(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteTaskRequest
	requestParams.ID = c.Param("id")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 删除任务
	err := jobHelper.DeleteJob(requestParams.ID)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除任务: %s - %s\n", requestParams.ID, err.Error())
		}
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, map[string]string{"message": "任务删除成功"})
}

// 启动调度器
func StartTask(c *gin.Context) {
	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 获取全部任务
	jobs := jobHelper.GetJobsData(5)
	// 启动 GoCron 调度器
	initialize.StartGoCron()
	if len(jobs) == 0 {
		// 启动CDN刷新任务
		ipUpdate := backend.NewIpUpdateService(global.GvaEntDb, global.GvaGoCron)
		if err := ipUpdate.StartCdnJob(c); err != nil {
			if global.GvaEnv.DEBUG {
				fmt.Printf("启动CDN刷新任务失败: %s\n", err.Error())
			}
			errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("启动CDN刷新任务失败: %v", err))
			return
		}
	}
	successResponse(c, map[string]string{"message": "调度器启动成功"})
}

// 停止调度器
func StopTask(c *gin.Context) {
	// 停止 GoCron 调度器
	initialize.StopGoCron()
	successResponse(c, map[string]string{"message": "调度器停止成功"})
}

// 清理全部任务
func CleanTask(c *gin.Context) {
	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 获取全部任务
	jobs := jobHelper.GetJobsData(5)
	// 清理任务
	for _, job := range jobs {
		err := jobHelper.DeleteJob(job.ID)
		if err != nil {
			if global.GvaEnv.DEBUG {
				fmt.Printf("清理全部任务: %s\n", err.Error())
			}
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}
	}
	successResponse(c, map[string]string{"message": "任务清理成功"})
}
