package main

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string            `bson:"username" json:"username"`
	Password  string            `bson:"password" json:"-"`
	IsAdmin   bool              `bson:"isAdmin" json:"isAdmin"`
	CreatedAt time.Time         `bson:"createdAt" json:"createdAt"`
}

type UserService struct {
	db *mongo.Database
}

func newUserService(db *mongo.Database) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(username, password string, isAdmin bool) (*User, error) {
	// Check if user already exists
	exists, _ := s.db.Collection("users").CountDocuments(context.Background(), bson.M{"username": username})
	if exists > 0 {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:        primitive.NewObjectID(),
		Username:  username,
		Password:  string(hashedPassword),
		IsAdmin:   isAdmin,
		CreatedAt: time.Now(),
	}

	_, err = s.db.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) AuthenticateUser(username, password string) (*User, error) {
	var user User
	err := s.db.Collection("users").FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func (s *UserService) ListUsers() ([]User, error) {
	cursor, err := s.db.Collection("users").Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []User
	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) DeleteUser(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := s.db.Collection("users").DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (s *UserService) UpdateUser(id string, username string, isAdmin bool) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"username": username,
			"isAdmin":  isAdmin,
		},
	}

	result, err := s.db.Collection("users").UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (s *UserService) ChangePassword(id string, newPassword string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"password": string(hashedPassword),
		},
	}

	result, err := s.db.Collection("users").UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}