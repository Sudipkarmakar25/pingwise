package v1

import (
	"github.com/labstack/echo/v4"

	"pingwise/cmd/api/handlers/v1/alerts"
	"pingwise/cmd/api/handlers/v1/auth"
	"pingwise/cmd/api/handlers/v1/services"
	"pingwise/cmd/api/handlers/v1/users"
	customMiddlewares "pingwise/cmd/api/middlewares"
	"pingwise/internal/models/enums"
)

func RegisterHandlers(v1 *echo.Group) {
	registerAuthHandlers(v1)
	registerUserHandlers(v1)
	registerServiceHandlers(v1)
}

func registerAuthHandlers(v1 *echo.Group) {
	authRoutes := v1.Group("/auth")

	authRoutes.POST("/signup/", auth.GetAuthHandler().Signup)
	authRoutes.POST("/login/", auth.GetAuthHandler().Login)
	authRoutes.POST("/logout/", auth.GetAuthHandler().Logout, customMiddlewares.AuthTokenMiddleware)
	authRoutes.POST("/change_password/", auth.GetAuthHandler().ChangePassword, customMiddlewares.AuthTokenMiddleware)
}

func registerUserHandlers(v1 *echo.Group) {
	userRoutes := v1.Group("/users", customMiddlewares.AuthTokenMiddleware)
	ownerOnly := customMiddlewares.RoleMiddleware(enums.UserRoleOwner)

	userRoutes.GET("/me/", users.GetUserHandler().GetMe)
	userRoutes.GET("/", users.GetUserHandler().ListUsers, ownerOnly)
	userRoutes.POST("/", users.GetUserHandler().AddUser, ownerOnly)
	userRoutes.PATCH("/:id/", users.GetUserHandler().UpdateUserRole, ownerOnly)
	userRoutes.DELETE("/:id/", users.GetUserHandler().RemoveUser, ownerOnly)
}

func registerServiceHandlers(v1 *echo.Group) {
	serviceRoutes := v1.Group("/services", customMiddlewares.AuthTokenMiddleware)
	canEdit := customMiddlewares.RoleMiddleware(enums.UserRoleOwner, enums.UserRoleEditor)

	serviceRoutes.GET("/", services.GetServiceHandler().ListServices)
	serviceRoutes.GET("/stats/", services.GetServiceHandler().ListStats)
	serviceRoutes.GET("/:id/", services.GetServiceHandler().GetService)
	serviceRoutes.GET("/:id/checks/", services.GetServiceHandler().ListChecks)
	serviceRoutes.GET("/:id/timeline/", services.GetServiceHandler().GetTimeline)
	serviceRoutes.POST("/", services.GetServiceHandler().CreateService, canEdit)
	serviceRoutes.PATCH("/:id/", services.GetServiceHandler().UpdateService, canEdit)
	serviceRoutes.DELETE("/:id/", services.GetServiceHandler().DeleteService, canEdit)

	// who gets alert emails; the permission rules are in the alert service
	serviceRoutes.GET("/:id/alerts/", alerts.GetAlertHandler().ListServiceAlerts)
	serviceRoutes.PATCH("/:id/alerts/:userId/", alerts.GetAlertHandler().UpdateServiceAlert)
}
