package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/oldlay/simplebank/db/sqlc"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	store  db.Store
	router *gin.Engine
}

// Create a new HTTP server and setup routing
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	router.POST("/users", server.createUser)

	// CREATE - POST /accounts
	router.POST("/accounts", server.createAccount)
	// READ - GET /accounts/:id
	router.GET("/accounts/:id", server.getAccount)
	// LIST - GET /accounts
	router.GET("/accounts", server.listAccounts)
	// UPDATE - PUT /accounts/:id (更新整个资源) 或 PATCH /accounts/:id (部分更新)
	router.PATCH("/accounts/:id", server.updateAccount)
	// DELETE - DELETE /accounts/:id
	router.DELETE("/accounts/:id", server.deleteAccount)

	router.POST("/transfers", server.createTransfer)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
