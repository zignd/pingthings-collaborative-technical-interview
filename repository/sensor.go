package repository

import (
	"context"

	"github.com/zignd/pingthings-collaborative-technical-interview/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Sensor struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name" json:"name"`
	Location GeoJSONPoint       `bson:"location" json:"location"`
	Tags     []string           `bson:"tags" json:"tags"`
}

type GeoJSONPoint struct {
	Type        string    `bson:"type" json:"type"`               // Should be "Point"
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // Longitude, Latitude
}

type SensorsRepository struct {
	mongoClient *mongo.Client
	sensorsColl *mongo.Collection
}

func NewSensorsRepository(envVars *config.EnvVars, mongoClient *mongo.Client) (*SensorsRepository, error) {
	sensorsColl := mongoClient.Database(envVars.MongoDB.Database).Collection("sensors")
	_, err := sensorsColl.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"location": "2dsphere",
		},
	})
	if err != nil {
		return nil, err
	}

	return &SensorsRepository{
		mongoClient: mongoClient,
		sensorsColl: sensorsColl,
	}, nil
}

func (s *SensorsRepository) Close() error {
	return s.mongoClient.Disconnect(context.Background())
}

func (s *SensorsRepository) CreateSensor(ctx context.Context, sensor *Sensor) error {
	result, err := s.sensorsColl.InsertOne(ctx, sensor)
	if err != nil {
		return err
	}
	sensor.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (s *SensorsRepository) GetSensorByID(ctx context.Context, id string) (*Sensor, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var sensor Sensor
	if err = s.sensorsColl.FindOne(ctx, bson.M{"_id": objectID}).Decode(&sensor); err != nil {
		return nil, err
	}
	return &sensor, nil
}

func (s *SensorsRepository) GetSensorByName(ctx context.Context, name string) (*Sensor, error) {
	var sensor Sensor
	if err := s.sensorsColl.FindOne(ctx, bson.M{"name": name}).Decode(&sensor); err != nil {
		return nil, err
	}
	return &sensor, nil
}

func (s *SensorsRepository) GetNearestSensor(ctx context.Context, latitude, longitude, maxDistance float64) (*Sensor, error) {
	var sensor Sensor
	if err := s.sensorsColl.FindOne(ctx, bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{longitude, latitude},
				},
				"$maxDistance": maxDistance,
			},
		},
	}).Decode(&sensor); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &sensor, nil
}

func (s *SensorsRepository) UpdateSensor(ctx context.Context, id string, sensor *Sensor) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = s.sensorsColl.UpdateByID(ctx, objectID, bson.M{
		"$set": bson.M{
			"name":     sensor.Name,
			"location": sensor.Location,
			"tags":     sensor.Tags,
		},
	})

	sensor.ID = objectID

	return err
}
