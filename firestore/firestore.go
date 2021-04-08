package firestore

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func Save() {
	ctx := context.Background()
	sa := option.WithCredentialsFile("path/to/serviceAccount.json")
	client := createClient(ctx, sa)
	defer client.Close()

	_, _, err := client.Collection("rfa").Add(ctx, map[string]interface{}{
		"name":  "First User",
		"age":   11,
		"email": "first@example.com",
	})
	if err != nil {
		fmt.Println(err)
	}
}

func createClient(ctx context.Context, sa option.ClientOption) *firestore.Client {
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	return client
}
