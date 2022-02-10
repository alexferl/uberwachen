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
	ReplicaSet             string
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

func (md *MongoDB) Init(ctx context.Context) (err error) {
	b := &backoff.Backoff{
		Jitter: true,
	}

	opts := options.Client()
	opts.SetSocketTimeout(md.opts.SocketTimeout)
	opts.SetServerSelectionTimeout(md.opts.ServerSelectionTimeout)
	opts.SetConnectTimeout(md.opts.ConnectTimeout)
	if md.opts.ReplicaSet != "" {
		opts.SetReplicaSet(md.opts.ReplicaSet)
	}

	clientOpts := opts.ApplyURI(md.opts.URI)
	if md.opts.Username != "" && md.opts.Password != "" {
		clientOpts.SetAuth(options.Credential{
			Username: md.opts.Username,
			Password: md.opts.Password,
		})
	}

	for {
		client, err := mongo.Connect(ctx, clientOpts)
		md.client = client
		if err != nil {
			d := b.Duration()
			log.Error().Msgf("%v, reconnecting in %s", err, d)
			time.Sleep(d)
			continue
		}

		b.Reset()

		md.c = client.Database(md.opts.DatabaseName).Collection(IncidentsColl)

		t := true
		_, err = md.c.Indexes().CreateOne(ctx, mongo.IndexModel{
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

func (md *MongoDB) Get(ctx context.Context, name string, doc interface{}) error {
	err := md.c.FindOne(ctx, bson.M{"name": name}).Decode(doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}

	return nil
}

func (md *MongoDB) GetAll(ctx context.Context, docs interface{}) error {
	cur, err := md.c.Find(ctx, bson.M{})
	if err != nil {
		return err
	}

	err = cur.All(ctx, docs)
	if err != nil {
		return err
	}

	return nil
}

func (md *MongoDB) Set(ctx context.Context, data interface{}) error {
	_, err := md.c.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (md *MongoDB) Update(ctx context.Context, name string, data interface{}) error {
	filter := bson.M{"name": name}
	update := bson.M{"$set": data}
	err := md.c.FindOneAndUpdate(ctx, filter, update).Decode(data)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}

	return nil
}

func (md *MongoDB) Delete(ctx context.Context, name string) error {
	_, err := md.c.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		return err
	}

	return nil
}
