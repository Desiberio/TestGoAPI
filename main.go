package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	cnnStr := os.Getenv("TestDB")
	fmt.Println(cnnStr)

	db, err := sqlx.Connect("sqlserver", cnnStr)
	if err != nil {
		log.Fatal("Error opening database connection", err)
	}
	defer db.Close()

	r := gin.Default()

	r.GET("/list", func(c *gin.Context) {

		var news []News
		err = db.Select(&news, "SELECT * FROM News")
		if err != nil {
			fmt.Println("Error retrieving news from the database:", err)
			return
		}

		for i := 0; i < len(news); i++ {
			err = db.Select(&news[i].Categories, "SELECT CategoryId as Categories FROM NewsCategories as NC WHERE NC.NewsId = @p1", news[i].Id)
			if err != nil {
				fmt.Println("Error retrieving categories from the database:", err)
				continue
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"Success": true,
			"news":    news,
		})
	})

	r.POST("/edit/:id", func(c *gin.Context) {
		id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Success":   false,
				"error":     "Incorrect Id.",
				"errorCode": 400,
			})
			return
		}

		var updatedNews News
		if err := c.ShouldBindJSON(&updatedNews); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Success":   false,
				"error":     "Bad Request",
				"errorCode": 400,
			})
			return
		}

		var count int
		err = db.Get(&count, "SELECT COUNT(Id) FROM News WHERE Id = @p1", id64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Success":   false,
				"error":     "Error while processing: " + err.Error(),
				"errorCode": 500,
			})
			return
		}

		if count > 1 || count == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"Success":   false,
				"error":     "There is no news with provided Id.",
				"errorCode": 404,
			})
			return
		}

		db.MustExec("UPDATE News SET Id = @p1, Title = @p2, Content = @p3 WHERE Id = @p4", updatedNews.Id, updatedNews.Title, updatedNews.Content, id64)
		//Кажется, у Reform была более интересная возможность управлять состояниями объекта в БД, поэтому она могла бы сама управлять состоянием массива Categories при изменениях.
		//sqlx же такой возможности, кажется, не имеет, поэтому я выбрал более простое и быстрое в разработке решение: удалять все записи из второй таблицы NewsCategories и добавлять их заново.

		db.MustExec("DELETE FROM NewsCategories WHERE NewsId = @p1", id64)
		for j := 0; j < len(updatedNews.Categories); j++ {
			db.MustExec("INSERT INTO NewsCategories VALUES (@p1, @p2)", updatedNews.Id, updatedNews.Categories[j])
		}

		//делаем преобразование, чтобы такие симовлы, как <, > и & отправлялись не в виде сырых кодов Unicode
		data, _ := updatedNews.JSON()
		c.Data(http.StatusOK, "application/json", data)
	})

	r.Run()
}
