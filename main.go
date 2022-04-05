package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// You will be using this Trainer type later in the program
type Url struct {
    Address string
}

type Mail struct {
    Sender  string
    To      []string
    Subject string
    Body    string
}

func main() {
	godotenv.Load(".env")

	from := os.Getenv("FROM_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	to := []string{os.Getenv("TO_EMAIL")}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	c := cron.New()

	fmt.Println("Started...")

	// TODO: Change cron expression here
	c.AddFunc("@every 1m", func() {
		fmt.Println("Running Cron Job")
		fmt.Println("Current time:", time.Now())
		
		clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
		client, err := mongo.Connect(context.TODO(), clientOptions)

		if err != nil {
			log.Fatal(err)
		}

		err = client.Ping(context.TODO(), nil)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Connected to MongoDB!")

		// TODO: Change based on database and collection name
		urlCollection := client.Database("myFirstDatabase").Collection("urls")

		var results []*Url
		cursor, err := urlCollection.Find(context.TODO(), bson.D{{}})

		for cursor.Next(context.TODO()) {
			var elem Url
			err := cursor.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}
		
			results = append(results, &elem)
		}

		if err != nil {
			log.Fatal(err)
		}

		cursor.Close(context.TODO())

		var wg sync.WaitGroup

		for _, item := range results {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				resp, err := http.Get(url)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(url, resp.StatusCode, http.StatusText(resp.StatusCode))

				title := fmt.Sprintf("%s %d %s", url, resp.StatusCode, http.StatusText(resp.StatusCode))

				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					fmt.Println("HTTP Status is in the 2xx range")
				} else {

					request := Mail{
						Sender:  from,
						To:      to,
						Subject: title,
						Body:    "",
					}
					message := BuildMessage(request)
					
					auth := smtp.PlainAuth("", from, password, smtpHost)
					err := smtp.SendMail(smtpHost + ":" + smtpPort, auth, from, to, []byte(message))
					if err != nil {
						log.Fatal(err)
					}
				}
			}(item.Address)
		}
		wg.Wait()
	})

	c.Start()
	select {}
}

func BuildMessage(mail Mail) string {
    msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
    msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
    msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
    msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
    msg += fmt.Sprintf("\r\n%s\r\n", mail.Body)

    return msg
}