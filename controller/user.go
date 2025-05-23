package controller

import (
	"encoding/json"
	"gin-template/common"
	"gin-template/model"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	if !common.PasswordLoginEnabled {
		ResponseError(c, "管理员关闭了密码登录")
		return
	}
	var loginRequest LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&loginRequest)
	if err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	username := loginRequest.Username
	password := loginRequest.Password
	if username == "" || password == "" {
		ResponseError(c, "无效的参数")
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}
	err = user.ValidateAndFill()
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	setupLogin(&user, c)
}

// setup session & cookies and then return user info
func setupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	err := session.Save()
	if err != nil {
		ResponseError(c, "无法保存会话信息，请重试")
		return
	}
	cleanUser := model.User{
		Id:          user.Id,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
	}
	ResponseOK(c, cleanUser)
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, nil)
}

func Register(c *gin.Context) {
	if !common.RegisterEnabled {
		ResponseError(c, "管理员关闭了新用户注册")
		return
	}
	if !common.PasswordRegisterEnabled {
		ResponseError(c, "管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册")
		return
	}
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		ResponseError(c, "输入不合法 "+err.Error())
		return
	}
	if common.EmailVerificationEnabled {
		if user.Email == "" || user.VerificationCode == "" {
			ResponseError(c, "管理员开启了邮箱验证，请输入邮箱地址和验证码")
			return
		}
		if !common.VerifyCodeWithKey(user.Email, user.VerificationCode, common.EmailVerificationPurpose) {
			ResponseError(c, "验证码错误或已过期")
			return
		}
	}
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.Username,
	}
	if common.EmailVerificationEnabled {
		cleanUser.Email = user.Email
	}
	if err := cleanUser.Insert(); err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, nil)
}

func GetAllUsers(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	users, err := model.GetAllUsers(p*common.ItemsPerPage, common.ItemsPerPage)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, users)
}

func SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	users, err := model.SearchUsers(keyword)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, users)
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	user, err := model.GetUserById(int64(id), false)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role {
		ResponseError(c, "无权获取同级或更高等级用户的信息")
		return
	}
	ResponseOK(c, user)
}

func GenerateToken(c *gin.Context) {
	id := c.GetInt64("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	user.Token = uuid.New().String()
	user.Token = strings.Replace(user.Token, "-", "", -1)

	if model.DB.Where("token = ?", user.Token).First(user).RowsAffected != 0 {
		ResponseError(c, "请重试，系统生成的 UUID 竟然重复了！")
		return
	}

	if err := user.Update(false); err != nil {
		ResponseError(c, err.Error())
		return
	}

	ResponseOK(c, user.Token)
}

func GetSelf(c *gin.Context) {
	id := c.GetInt64("id")
	user, err := model.GetUserById(id, false)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, user)
}

func UpdateUser(c *gin.Context) {
	var updatedUser model.User
	err := json.NewDecoder(c.Request.Body).Decode(&updatedUser)
	if err != nil || updatedUser.Id == 0 {
		ResponseError(c, "无效的参数")
		return
	}
	if updatedUser.Password == "" {
		updatedUser.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&updatedUser); err != nil {
		ResponseError(c, "输入不合法 "+err.Error())
		return
	}
	originUser, err := model.GetUserById(updatedUser.Id, false)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role {
		ResponseError(c, "无权更新同权限等级或更高权限等级的用户信息")
		return
	}
	if myRole <= updatedUser.Role {
		ResponseError(c, "无权将其他用户权限等级提升到大于等于自己的权限等级")
		return
	}
	if updatedUser.Password == "$I_LOVE_U" {
		updatedUser.Password = "" // rollback to what it should be
	}
	updatePassword := updatedUser.Password != ""
	if err := updatedUser.Update(updatePassword); err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, nil)
}

func UpdateSelf(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	if user.Password == "" {
		user.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&user); err != nil {
		ResponseError(c, "输入不合法 "+err.Error())
		return
	}

	cleanUser := model.User{
		Id:          c.GetInt64("id"),
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if user.Password == "$I_LOVE_U" {
		user.Password = "" // rollback to what it should be
		cleanUser.Password = ""
	}
	updatePassword := user.Password != ""
	if err := cleanUser.Update(updatePassword); err != nil {
		ResponseError(c, err.Error())
		return
	}

	ResponseOK(c, nil)
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	originUser, err := model.GetUserById(int64(id), false)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role {
		ResponseError(c, "无权删除同权限等级或更高权限等级的用户")
		return
	}
	err = model.DeleteUserById(int64(id))
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, nil)
}

func DeleteSelf(c *gin.Context) {
	id := c.GetInt64("id")
	err := model.DeleteUserById(id)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, nil)
}

func CreateUser(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" {
		ResponseError(c, "无效的参数")
		return
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	myRole := c.GetInt("role")
	if user.Role >= myRole {
		ResponseError(c, "无法创建权限大于等于自己的用户")
		return
	}
	// Even for admin users, we cannot fully trust them!
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if err := cleanUser.Insert(); err != nil {
		ResponseError(c, err.Error())
		return
	}

	ResponseOK(c, nil)
}

type ManageRequest struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

// ManageUser Only admin user can do this
func ManageUser(c *gin.Context) {
	var req ManageRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)

	if err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	user := model.User{
		Username: req.Username,
	}
	// Fill attributes
	model.DB.Where(&user).First(&user)
	if user.Id == 0 {
		ResponseError(c, "用户不存在")
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != common.RoleRootUser {
		ResponseError(c, "无权更新同权限等级或更高权限等级的用户信息")
		return
	}
	switch req.Action {
	case "disable":
		user.Status = common.UserStatusDisabled
		if user.Role == common.RoleRootUser {
			ResponseError(c, "无法禁用超级管理员用户")
			return
		}
	case "enable":
		user.Status = common.UserStatusEnabled
	case "delete":
		if user.Role == common.RoleRootUser {
			ResponseError(c, "无法删除超级管理员用户")
			return
		}
		if err := user.Delete(); err != nil {
			ResponseError(c, err.Error())
			return
		}
	case "promote":
		if myRole != common.RoleRootUser {
			ResponseError(c, "普通管理员用户无法提升其他用户为管理员")
			return
		}
		if user.Role >= common.RoleAdminUser {
			ResponseError(c, "该用户已经是管理员")
			return
		}
		user.Role = common.RoleAdminUser
	case "demote":
		if user.Role == common.RoleRootUser {
			ResponseError(c, "无法降级超级管理员用户")
			return
		}
		if user.Role == common.RoleCommonUser {
			ResponseError(c, "该用户已经是普通用户")
			return
		}
		user.Role = common.RoleCommonUser
	}

	if err := user.Update(false); err != nil {
		ResponseError(c, err.Error())
		return
	}
	clearUser := model.User{
		Role:   user.Role,
		Status: user.Status,
	}
	ResponseOK(c, clearUser)
}

func EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")
	if !common.VerifyCodeWithKey(email, code, common.EmailVerificationPurpose) {
		ResponseError(c, "验证码错误或已过期")
		return
	}
	id := c.GetInt64("id")
	user := model.User{
		Id: id,
	}
	err := user.FillUserById()
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	user.Email = email
	// no need to check if this email already taken, because we have used verification code to check it
	err = user.Update(false)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	ResponseOK(c, nil)
}
