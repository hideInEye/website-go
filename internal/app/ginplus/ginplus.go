package ginplus

import (
	"context"
	stderrors "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"net/http"
	"strings"

	"go-gin/internal/app/config"

	icontext "go-gin/internal/app/context"
	"go-gin/internal/app/errors"
	"go-gin/internal/app/schema"
	"go-gin/pkg/logger"
	"go-gin/until"
)

// 定义上下文中的键
const (
	prefix = "ginadmin"
	// UserIDKey 存储上下文中的键(用户ID)
	UserIDKey = prefix + "/user_id"
	// TraceIDKey 存储上下文中的键(跟踪ID)
	TraceIDKey = prefix + "/trace_id"
	// ResBodyKey 存储上下文中的键(响应Body数据)
	ResBodyKey = prefix + "/res_body"
	// AppParameterKey 存储上下文中的键(APP参数)
	AppParameterKey = prefix + "/app_parameter"
	//MenuIDKey 存储上下文中的键(菜单ID)
	MenuIDKey = prefix + "/menu_id"
)

// NewContext 封装上线文入口
func NewContext(c *gin.Context) context.Context {
	parent := c.Request.Context()

	traceId := GetTraceID(c)

	parent = icontext.NewTraceID(parent, traceId)
	parent = logger.NewTraceIDContext(parent, traceId)

	if v := GetUserID(c); v != "" {
		parent = icontext.NewUserID(parent, v)
		parent = logger.NewUserIDContext(parent, v)
	}

	return parent
}

// GetToken 获取用户令牌
func GetToken(c *gin.Context) string {
	var token string
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "
	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	}

	if token == "" {
		token = c.Query("token")
	}
	return token
}

// GetPageIndex 获取分页的页索引
func GetPageIndex(c *gin.Context) int {
	defaultVal := 1
	if v := c.Query("current"); v != "" {
		if iv := until.S(v).DefaultInt(defaultVal); iv > 0 {
			return iv
		}
	}
	return defaultVal
}

// GetPageSize 获取分页的页大小(最大50)
func GetPageSize(c *gin.Context) int {
	defaultVal := 10
	if v := c.Query("pageSize"); v != "" {
		if iv := until.S(v).DefaultInt(defaultVal); iv > 0 {
			if iv > 1000 {
				iv = 1000
			}
			return iv
		}
	}
	return defaultVal
}

// GetPaginationParam 获取分页查询参数
func GetPaginationParam(c *gin.Context) *schema.PaginationParam {
	return &schema.PaginationParam{
		PageIndex: GetPageIndex(c),
		PageSize:  GetPageSize(c),
	}
}

// GetTraceID 获取追踪ID
func GetTraceID(c *gin.Context) string {
	return c.GetString(TraceIDKey)
}

// GetUserID 获取用户ID
func GetUserID(c *gin.Context) string {
	return c.GetString(UserIDKey)
}

func GetMenuID(c *gin.Context) string {
	return c.GetString(MenuIDKey)
}

// CheckIsRootUser 检查用户是否是root用户
func CheckIsRootUser(c *gin.Context) bool {
	return config.GetGlobalConfig().Root.UserName == GetUserID(c)
}

// SetUserID 设定用户ID
func SetUserID(c *gin.Context, userID string) {
	c.Set(UserIDKey, userID)
}

// ParseJSON 解析请求JSON
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		logger.Warnf(NewContext(c), "%v", err)
		return errors.ErrInvalidRequestParameter
	}
	return nil
}

// ResPage 响应分页数据
func ResPage(c *gin.Context, v interface{}, pr *schema.PaginationResult) {
	list := schema.HTTPList{
		List: v,
		Pagination: &schema.HTTPPagination{
			Current:  GetPageIndex(c),
			PageSize: GetPageSize(c),
		},
	}
	if pr != nil {
		list.Pagination.Total = pr.Total
	}

	ResSuccess(c, list)
}

// ResList 响应列表数据
func ResList(c *gin.Context, v interface{}) {
	ResSuccess(c, schema.HTTPList{List: v})
}

// ResOK 响应OK
func ResOK(c *gin.Context) {
	ResSuccess(c, schema.HTTPStatus{Status: "OK"})
}

// ResSuccess 响应成功
func ResSuccess(c *gin.Context, v interface{}) {
	ResJSON(c, http.StatusOK, v)
}

// ResJSON 响应JSON数据
func ResJSON(c *gin.Context, status int, v interface{}) {
	buf, err := until.JSONMarshal(v)
	if err != nil {
		panic(err)
	}
	c.Set(ResBodyKey, buf)
	c.Data(status, "application/json; charset=utf-8", buf)
	c.Abort()
}

// ResError 响应错误
func ResError(c *gin.Context, err error, status ...int) {
	statusCode := http.StatusInternalServerError
	message := "数据操作失败"
	var publicError *errors.HTTPError
	var mysqlError *mysql.MySQLError
	if stderrors.As(err, &publicError) {
		statusCode = publicError.Status
		message = publicError.Message
	} else if gorm.IsRecordNotFoundError(err) {
		statusCode = http.StatusNotFound
		message = "记录不存在"
	} else if stderrors.As(err, &mysqlError) && (mysqlError.Number == 1062 || mysqlError.Number == 1451 || mysqlError.Number == 1452) {
		statusCode = http.StatusConflict
		message = "数据重复或仍被其他记录使用"
	} else if errCode, ok := errors.FromErrorCode(err); ok {
		statusCode = errCode.HTTPStatusCode
		message = errCode.Message
	}

	if len(status) > 0 {
		statusCode = status[0]
	}

	if statusCode == 500 && err != nil {
		span := logger.StartSpan(NewContext(c))
		span = span.WithField("stack", fmt.Sprintf("%+v", err))
		span.Errorf("%v", err)
	}

	Fail(c, statusCode, message)
}

// ResData preserves the API response envelope consumed by the existing frontend.
func ResData(c *gin.Context, status int, data interface{}) {
	ResJSON(c, status, schema.Object{"data": data})
}

func Fail(c *gin.Context, status int, message string) {
	ResJSON(c, status, schema.Object{
		"statusCode": status,
		"code":       fmt.Sprintf("HTTP_%d", status),
		"message":    message,
		"traceId":    GetTraceID(c),
	})
}

func Bind(c *gin.Context, value interface{}) bool {
	if err := c.ShouldBindJSON(value); err != nil {
		Fail(c, http.StatusBadRequest, "请求参数格式或内容不正确")
		return false
	}
	return true
}

func SetCurrentUser(c *gin.Context, user *schema.User) {
	c.Set(prefix+"/current_user", user)
	SetUserID(c, user.ID)
}

func CurrentUser(c *gin.Context) *schema.User {
	value, _ := c.Get(prefix + "/current_user")
	user, _ := value.(*schema.User)
	return user
}
