package app

import (
	v1Handler "chat-app/internal/handlers/v1"
	"chat-app/internal/repository"
	"chat-app/internal/routes"
	v1Routes "chat-app/internal/routes/v1"
	services "chat-app/internal/services/v1"
	"chat-app/pkg/auth"
	"chat-app/pkg/cache"
)

type AuthModule struct {
	routes routes.Routes
}

func NewAuthModule(ctx *ModuleContext, tokenService auth.TokenService, cache cache.RedisCacheService) *AuthModule {
	// init repositories
	userRepo := repository.NewSqlUserRepository(ctx.DB)
	// TokenService auth.TokenService, cacheService cache.RedisCacheService

	// init services
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, tokenService, cache)

	// init handlers
	authHandler := v1Handler.NewAuthHandler(userService, authService , tokenService)

	// init routes
	authRoutes := v1Routes.NewAuthRoutes(authHandler)

	return &AuthModule{
		routes: authRoutes,
	}
}

func (am *AuthModule) GetRoutes() routes.Routes {
	return am.routes
}
