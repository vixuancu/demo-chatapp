package app

import (
	"chat-app/internal/config"
	"chat-app/internal/db"
	"chat-app/internal/db/sqlc"
	"chat-app/internal/routes"
	"chat-app/internal/validation"
	"chat-app/pkg/websocket"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Module interface {
	GetRoutes() routes.Routes
}
type Application struct {
	config    *config.Config
	router    *gin.Engine
	modules   []Module
	wsManager *websocket.Manager
}
type ModuleContext struct {
	DB sqlc.Querier
	WSManager *websocket.Manager
}

func NewApplication(cfg *config.Config) *Application {
	r := gin.Default()
	if err := validation.InitValidator(); err != nil {
		log.Fatal("cannot initialize validator:", err)
	}
	if err := db.InitDB(); err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	// Create and start WebSocket manager
	wsManager := websocket.NewManager()
	go wsManager.Run()


	ctx := &ModuleContext{
		DB: db.DB,
		WSManager: wsManager,
	}
	modules := []Module{
		NewUserModule(ctx),
		// NewAuthModule(ctx),
		NewRoomModule(ctx), // thêm module Room
		NewChatModule(ctx),
	}
	routes.RegisterRoutes(r, GetModuleRoutes(modules)...)

	return &Application{
		config:  cfg,
		router:  r,
		modules: modules,
		wsManager: wsManager,
	}
}

// hàm lấy tất cả routes từ các module
func GetModuleRoutes(modules []Module) []routes.Routes {
	routesList := make([]routes.Routes, len(modules))
	for i, module := range modules {
		routesList[i] = module.GetRoutes()
	}
	return routesList
}
func (app *Application) Run() error {
	// if err := app.router.Run(app.config.ServerAddress); err != nil {
	// 	return err
	// }
	// comment bằng Tiếng Việt
	svr := &http.Server{
		Addr:    app.config.ServerAddress,
		Handler: app.router,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP) // khi nhấn Ctrl+C hoặc dừng server hoắc reload

	// Chạy server trong một goroutine vì để tránh blocking

	go func() {
		log.Printf("❤️ Starting server on %s", app.config.ServerAddress)
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ ListenAndServe failed: %v", err)
		}
	}()

	<-quit // Chờ tín hiệu dừng
	log.Println("🍺 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		log.Fatalf("⚠️ Server forced to shutdown: %v", err)
	}
	log.Println("🍺 Server exiting")
	return nil
}
