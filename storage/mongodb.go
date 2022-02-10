package storage

import (
	"context"
	"time"

	"github.com/jpillora/backoff"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	IncidentsColl = "incidents"
)

type MongoDBOpts struct {
	URI                    string
	DatabaseName           string
	Username               string
	Password               string
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
	SocketTimeout          time.Duration
}

type MongoDB struct {
	client *mongo.Client
	c      *mongo.Collection
	opts   *MongoDBOpts
}

func NewMongoDB(opts *MongoDBOpts) (Storage, error) {
	return Storage(&MongoDB{opts: opts}), nil
}

func (mb *MongoDB) Init(ctx context.Context) (err error) {
	b := &backoff.Backoff{
		Jitter: true,
	}

	opts := options.Client()
	opts.SetSocketTimeout(mb.opts.SocketTimeout)
	opts.SetServerSelectionTimeout(mb.opts.ServerSelectionTimeout)
	opts.SetConnectTimeout(mb.opts.ConnectTimeout)

	clientOpts := opts.ApplyURI(mb.opts.URI)
	if mb.opts.Username != "" && mb.opts.Password != "" {
		clientOpts.SetAuth(options.Credential{
			Username: mb.opts.Username,
			Password: mb.opts.Password,
		})
	}

	for {
		client, err := mongo.Connect(ctx, clientOpts)
		mb.client = client
		if err != nil {
			d := b.Duration()
			log.Error().Msgf("%v, reconnecting in %s", err, d)
			time.Sleep(d)
			continue
		}

		b.Reset()

		mb.c = client.Database(mb.opts.DatabaseName).Collection(IncidentsColl)

		t := true
		_, err = mb.c.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{"name", 1}},
			Options: &options.IndexOptions{
				Unique: &t,
			},
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func (mb *MongoDB) Get(ctx context.Context, name string, doc interface{}) error {
	err := mb.c.FindOne(ctx, bson.M{"name": name}).Decode(doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}

	return nil
}

func (mb *MongoDB) GetAll(ctx context.Context, docs interface{}) error {
	cur, err := mb.c.Find(ctx, bson.M{})
	if err != nil {
		return err
	}

	err = cur.All(ctx, docs)
	if err != nil {
		return err
	}

	return nil
}

func (mb *MongoDB) Set(ctx context.Context, data interface{}) error {
	_, err := mb.c.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (mb *MongoDB) Update(ctx context.Context, name string, data interface{}) error {
	filter := bson.M{"name": name}
	update := bson.M{"$set": data}
	err := mb.c.FindOneAndUpdate(ctx, filter, update).Decode(data)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}

	return nil
}

func (mb *MongoDB) Delete(ctx context.Context, name string) error {
	_, err := mb.c.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		return err
	}

	return nil
}
