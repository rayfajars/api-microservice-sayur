package handler

import (
	"net/http"
	"user-service/config"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type UserHandlerInterface interface {
	SignIn(ctx echo.Context) error
}

type userHandler struct {
	userService service.UserServiceInterface
}

var err error

func (u *userHandler) SignIn(c echo.Context) error {
	var (
		req        = request.SignInRequest{}
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if err = c.Validate(&req); err != nil {
		log.Errorf("[UserHandler-2] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := u.userService.SignIn(ctx, reqEntity)

	if err != nil {
		if err.Error() == "404" {
			log.Infof("[UserHandler-3] SignIn: %s", "User not found")
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		log.Errorf("[UserHandler-4] SignIn: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respSignIn = response.SignInResponse{
		AccessToken: token,
		Role:        user.RoleName,
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Phone:       user.Phone,
		Lat:         user.Lat,
		Lng:         user.Lng,
	}

	resp.Message = "Success"
	resp.Data = respSignIn
	return c.JSON(http.StatusOK, resp)
}

func NewUserHandler(e *echo.Echo, userService service.UserServiceInterface, cfg *config.Config) UserHandlerInterface {
	userHandler := &userHandler{userService: userService}

	e.Use(middleware.Recover())
	e.POST("/signin", userHandler.SignIn)

	mid := adapter.NewMiddlewareAdapter(cfg)

	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/check", func(c echo.Context) error {
		// return c.JSON(http.StatusOK, response.DefaultResponse{
		// 	Message: "Success",
		// 	Data:    "You are logged in",
		// })

		return c.String(200, "OK")
	})

	return userHandler
}
