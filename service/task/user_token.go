package task

import (
	"encoding/json"
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"gin-template/service"
)

func CompensationUserTokensInit(task model.CompensationTask) error {
	var params struct {
		userId int64 `json:"user_id"`
		amount int64 `json:"amount"`
	}

	if err := json.Unmarshal([]byte(task.Payload), &params); err != nil {
		return fmt.Errorf("参数解析失败: %w", err)
	}

	// 先检查是否已经初始化成功
	//if service.GetTokenService().IsInitialized(params.userId) {
	//	return nil // 已经成功则直接返回
	//}

	//todo 初始化失败，流水记录状态要设置为失败，在恢复的时候，要先检查流水表中失败的，随后对这条记录修复重试
	_, err := service.GetTokenService().InitUserTokenAccount(params.userId, params.amount)
	if err != nil {
		common.SysError(fmt.Sprintf("init user %d token account failed: %v", params.userId, err))
		return fmt.Errorf("初始化失败: %w", err)
	}

	return nil

}
