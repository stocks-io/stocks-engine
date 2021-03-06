package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type orderRequest struct {
	Token  string `form:"token" json:"token" binding:"required"`
	Units  int    `form:"units" json:"units" binding:"required"`
	Symbol string `form:"symbol" json:"symbol" binding:"required"`
}

type ownedRequest struct {
	Token string `form:"token" json:"token" binding:"required"`
}

type stockHistoryRequest struct {
	Token     string `form:"token" json:"token" binding:"required"`
	Timeframe string `form:"timeframe" json:"timeframe" binding:"required"`
	Symbol    string `form:"symbol" json:"symbol" binding:"required"`
}

func setupPortfolioRoutes() {
	portfolio := app.Group("/portfolio")
	{
		portfolio.POST("/buy", func(c *gin.Context) {
			var req orderRequest
			c.ShouldBindWith(&req, binding.Form)
			if req.Units < 0 {
				c.JSON(401, gin.H{"message": "Cannot buy negative units"})
				return
			}
			userId := tokenToUserId(req.Token)
			if userId == "" {
				c.JSON(401, gin.H{"message": "Unauthorized"})
				return
			}
			cash := getCash(userId)
			currentPrice, err := getStockPrice(req.Symbol)
			if err != nil {
				c.JSON(400, gin.H{"message": err.Error()})
				return
			}

			total := currentPrice * float64(req.Units)
			if cash < total {
				c.JSON(401, gin.H{"message": "Not enough money to buy"})
				return
			}

			cash -= total
			err = setCash(userId, cash)
			checkErr(err)
			updateUnitsOwned(userId, req, true)
			createOrder(userId, req, currentPrice, 1)
			c.JSON(200, gin.H{
				"message":       "Successfully ordered stocks",
				"totalCost":     total,
				"remainingCash": cash,
			})
		})
		portfolio.POST("/sell", func(c *gin.Context) {
			var req orderRequest
			c.ShouldBindWith(&req, binding.Form)
			if req.Units < 0 {
				c.JSON(401, gin.H{"message": "Cannot sell negative units"})
				return
			}

			userId := tokenToUserId(req.Token)
			if userId == "" {
				c.JSON(401, gin.H{"message": "Unauthorized"})
				return
			}

			unitsOwned := getUnitsOwned(userId, req.Symbol)
			if req.Units > unitsOwned {
				c.JSON(401, gin.H{"message": "Not enough units to sell"})
				return
			}

			cash := getCash(userId)
			currentPrice, err := getStockPrice(req.Symbol)
			checkErr(err)
			total := currentPrice * float64(req.Units)
			cash += total
			err = setCash(userId, cash)
			updateUnitsOwned(userId, req, false)
			createOrder(userId, req, currentPrice, 0)
			c.JSON(200, gin.H{
				"message":       "Successfully sold stocks",
				"totalCost":     total,
				"remainingCash": cash,
			})
		})

		portfolio.POST("/owned", func(c *gin.Context) {
			var req ownedRequest
			c.ShouldBindWith(&req, binding.Form)
			userId := tokenToUserId(req.Token)
			if userId == "" {
				c.JSON(401, gin.H{"message": "Unauthorized"})
				return
			}
			positions := getPositions(userId)
			c.JSON(200, positions)
		})

		portfolio.GET("/update/:userID", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"page": fmt.Sprintf("/update/%s", c.Param("userID")),
			})
		})

		portfolio.GET("/symbols", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"count":   len(symbols),
				"results": symbols,
			})
		})

		portfolio.GET("/stockhistory", func(c *gin.Context) {
			var req stockHistoryRequest
			c.Bind(&req)
			if req.Token == "" || req.Symbol == "" || req.Timeframe == "" {
				c.JSON(401, gin.H{"message": "token, symbol, and timeframe are all required"})
				return
			}
			userId := tokenToUserId(req.Token)
			if userId == "" {
				c.JSON(401, gin.H{"message": "Unauthorized"})
				return
			}
			allHistory, err := getStockHistory(req.Symbol, req.Timeframe)
			if err != nil {
				c.JSON(403, gin.H{
					"message": "getStockHistory failed",
					"error":   err.Error(),
				})
				return
			}
			c.JSON(200, gin.H{
				"count":   len(allHistory),
				"results": allHistory,
			})
		})
	}
}
