package adapter

import (
	"net/http"
	"strings"

	"user-service/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type MiddlewareAdapterInterface interface {
	CheckToken() echo.MiddlewareFunc
}

type middlewareAdapter struct {
	cfg *config.Config
}

// CheckToken implements [MiddlewareAdapterInterface].
func (m *middlewareAdapter) CheckToken() echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			redisConn := config.NewRedisClient()
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[MiddlewareAdapter-1] CheckToken: missing or invalid authorization header")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid authorization header")
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			getSession, err := redisConn.HGetAll(c.Request().Context(), tokenString).Result()
			if err != nil || len(getSession) == 0 {
				log.Errorf("[MiddlewareAdapter-2] CheckToken: %s", err.Error())
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid or expired token")
			}

			c.Set("user", getSession)
			return next(c)

		}
	}
}

func NewMiddlewareAdapter(cfg *config.Config) MiddlewareAdapterInterface {
	return &middlewareAdapter{
		cfg: cfg,
	}
}
