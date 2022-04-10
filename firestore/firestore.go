package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Participant struct{
	ID string
	Active bool `firestore:"active"`
}

func GetParticipants(ctx context.Context, projectID string) (participants []Participant, err error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return
	}

	iter := client.Collection("rfa-participants").Documents(ctx)
	for {
		var participant Participant

		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return participants, err
		}
		
		if err = doc.DataTo(&participant); err != nil {
			return participants, err
		}
		participant.ID = doc.Ref.ID
		if participant.Active {
			participants = append(participants, participant)
		}
	}

	return
}
