package web

import (
	"net/http"
	"time"

	jwt3 "github.com/a1270107629/studyroom/sr/bff/web/jwt"
	"github.com/a1270107629/studyroom/sr/pkg/ginx"
	userv1 "github.com/a1270107629/studyroom/sr/proto/gen/user"
	"github.com/a1270107629/studyroom/sr/user/errs"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`

	userIdKey = "userId"
	bizLogin  = "login"
)

var _ handler = &UserHandler{}

type UserHandler struct {
	svc              userv1.UserServiceClient
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	jwt3.Handler
}

func NewUserHandler(svc userv1.UserServiceClient, jwthdl jwt3.Handler) *UserHandler {
	return &UserHandler{
		svc:              svc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		Handler:          jwthdl,
	}
}

func (c *UserHandler) RegisterRoutes(server *gin.Engine) {

	// 分组注册
	ug := server.Group("/users")
	ug.POST("/signup", ginx.WrapReq[SignUpReq](c.SignUp))
	// session 机制
	//ug.POST("/login", c.Login)
	// JWT 机制
	ug.POST("/login", ginx.WrapReq[LoginReq](c.LoginJWT))
	ug.POST("/logout", c.Logout)
	ug.POST("/edit", c.Edit)
	//ug.GET("/profile", c.Profile)
	ug.GET("/profile", c.ProfileJWT)
	ug.POST("/refresh_token", c.RefreshToken)
}

func (c *UserHandler) RefreshToken(ctx *gin.Context) {
	// 假定长 token 也放在这里
	tokenStr := c.ExtractTokenString(ctx)
	var rc jwt3.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return jwt3.RefreshTokenKey, nil
	})
	// 这边要保持和登录校验一直的逻辑，即返回 401 响应
	if err != nil || token == nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, Result{Code: 4, Msg: "请登录"})
		return
	}

	// 校验 ssid
	err = c.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// 系统错误或者用户已经主动退出登录了
		// 这里也可以考虑说，如果在 Redis 已经崩溃的时候，
		// 就不要去校验是不是已经主动退出登录了。
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = c.SetJWTToken(ctx, rc.Ssid, rc.Id)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, Result{Code: 4, Msg: "请登录"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "刷新成功"})
}

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

// SignUp 用户注册接口
func (c *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {

	isEmail, err := c.emailRegexExp.MatchString(req.Email)
	if err != nil {
		return Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isEmail {
		return Result{
			Code: errs.UserInvalidInput,
			Msg:  "邮箱输入错误",
		}, nil
	}

	if req.Password != req.ConfirmPassword {
		return Result{
			Code: errs.UserInvalidInput,
			Msg:  "两次输入密码不对",
		}, nil
	}

	// isPassword, err := c.passwordRegexExp.MatchString(req.Password)
	// if err != nil {
	// 	return Result{
	// 		Code: errs.UserInvalidInput,
	// 		Msg:  "系统错误",
	// 	}, err
	// }
	// if !isPassword {
	// 	return Result{
	// 		Code: errs.UserInvalidInput,
	// 		Msg:  "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
	// 	}, nil
	// }

	_, err = c.svc.Signup(ctx.Request.Context(), &userv1.SignupRequest{User: &userv1.User{Email: req.Email, Password: req.ConfirmPassword}})
	if err != nil {
		return Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return Result{
		Msg: "OK",
	}, nil
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginJWT 用户登录接口，使用的是 JWT，如果你想要测试 JWT，就启用这个
func (c *UserHandler) LoginJWT(ctx *gin.Context, req LoginReq) (ginx.Result, error) {
	u, err := c.svc.Login(ctx.Request.Context(), &userv1.LoginRequest{
		Email: req.Email, Password: req.Password,
	})

	if err != nil {
		return ginx.Result{}, err
	}
	err = c.SetLoginToken(ctx, u.User.Id)
	if err != nil {
		return ginx.Result{}, err
	}
	return ginx.Result{Msg: "登录成功"}, nil
}

func (c *UserHandler) Logout(ctx *gin.Context) {
	err := c.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

// Login 用户登录接口
func (c *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	// 当我们调用 Bind 方法的时候，如果有问题，Bind 方法已经直接写响应回去了
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := c.svc.Login(ctx.Request.Context(), &userv1.LoginRequest{
		Email: req.Email, Password: req.Password})
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// TODO 利用 grpc 来传递错误码
	//if err == service.ErrInvalidUserOrPassword {
	//	ctx.String(http.StatusOK, "用户名或者密码不正确，请重试")
	//	return
	//}
	sess := sessions.Default(ctx)
	sess.Set(userIdKey, u.User.Id)
	sess.Options(sessions.Options{
		// 60 秒过期
		MaxAge: 60,
	})
	err = sess.Save()
	if err != nil {
		ctx.String(http.StatusOK, "服务器异常")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

// Edit 用户编译信息
func (c *UserHandler) Edit(ctx *gin.Context) {
	type Req struct {
		// 注意，其它字段，尤其是密码、邮箱和手机，
		// 修改都要通过别的手段
		// 邮箱和手机都要验证
		// 密码更加不用多说了
		Nickname string `json:"nickname"`
		// 2023-01-01
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 你可以尝试在这里校验。
	// 比如说你可以要求 Nickname 必须不为空
	// 校验规则取决于产品经理
	if req.Nickname == "" {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "昵称不能为空"})
		return
	}

	if len(req.AboutMe) > 1024 {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "关于我过长"})
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		// 也就是说，我们其实并没有直接校验具体的格式
		// 而是如果你能转化过来，那就说明没问题
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "日期格式不对"})
		return
	}

	uc := ctx.MustGet("user").(jwt3.UserClaims)
	_, err = c.svc.UpdateNonSensitiveInfo(ctx,
		&userv1.UpdateNonSensitiveInfoRequest{
			User: &userv1.User{
				Id:       uc.Id,
				Nickname: req.Nickname,
				AboutMe:  req.AboutMe,
				Birthday: timestamppb.New(birthday),
			},
		})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, Result{Msg: "OK"})
}

// ProfileJWT 用户详情, JWT 版本
func (c *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		Nickname string
		Birthday string
		AboutMe  string
	}
	uc := ctx.MustGet("user").(jwt3.UserClaims)
	resp, err := c.svc.Profile(ctx, &userv1.ProfileRequest{Id: uc.Id})
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	u := resp.User
	ctx.JSON(http.StatusOK, Profile{
		Email:    u.Email,
		Phone:    u.Phone,
		Nickname: u.Nickname,
		Birthday: u.Birthday.AsTime().Format(time.DateOnly),
		AboutMe:  u.AboutMe,
	})
}

// Profile 用户详情
func (c *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email string
	}
	sess := sessions.Default(ctx)
	id := sess.Get(userIdKey).(int64)
	u, err := c.svc.Profile(ctx, &userv1.ProfileRequest{
		Id: id,
	})
	if err != nil {
		// 按照道理来说，这边 id 对应的数据肯定存在，所以要是没找到，
		// 那就说明是系统出了问题。
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Email: u.User.Email,
	})
}
