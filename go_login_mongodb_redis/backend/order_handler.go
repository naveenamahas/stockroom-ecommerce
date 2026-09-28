package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func shopProductsHandler(w http.ResponseWriter, r *http.Request) {
	_, ctx, cancel, ok := sessionUserID(w, r)
	if !ok {
		return
	}
	defer cancel()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cursor, err := products.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "could not load shop products")
		return
	}
	defer cursor.Close(ctx)
	items := []Product{}
	if err = cursor.All(ctx, &items); err != nil {
		errorJSON(w, http.StatusInternalServerError, "could not read shop products")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ctx, cancel, ok := sessionUserID(w, r)
	if !ok {
		return
	}
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		cursor, err := orders.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, "could not load orders")
			return
		}
		defer cursor.Close(ctx)
		items := []Order{}
		if err = cursor.All(ctx, &items); err != nil {
			errorJSON(w, http.StatusInternalServerError, "could not read orders")
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		createOrder(w, r, userID, ctx)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func createOrder(w http.ResponseWriter, r *http.Request, userID bson.ObjectID, ctx context.Context) {
	var request PurchaseRequest
	if json.NewDecoder(r.Body).Decode(&request) != nil {
		errorJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	productID, err := bson.ObjectIDFromHex(strings.TrimSpace(request.ProductID))
	if err != nil || request.Quantity < 1 || request.PaymentMethod != "dummy" {
		errorJSON(w, http.StatusBadRequest, "product, positive quantity and dummy payment are required")
		return
	}

	var product Product
	err = products.FindOneAndUpdate(
		ctx,
		bson.M{"_id": productID, "quantity": bson.M{"$gte": request.Quantity}},
		bson.M{"$inc": bson.M{"quantity": -request.Quantity}, "$set": bson.M{"updatedAt": time.Now()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&product)
	if err == mongo.ErrNoDocuments {
		errorJSON(w, http.StatusConflict, "product is out of stock or has insufficient quantity")
		return
	}
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, "could not reserve product stock")
		return
	}

	order := Order{
		UserID:        userID,
		Items:         []OrderItem{{ProductID: product.ID, Name: product.Name, Price: product.Price, Quantity: request.Quantity}},
		Total:         product.Price * float64(request.Quantity),
		PaymentMethod: "dummy",
		Status:        "paid",
		CreatedAt:     time.Now(),
	}
	result, err := orders.InsertOne(ctx, order)
	if err != nil {
		_, _ = products.UpdateOne(ctx, bson.M{"_id": product.ID}, bson.M{"$inc": bson.M{"quantity": request.Quantity}})
		errorJSON(w, http.StatusInternalServerError, "could not create order")
		return
	}
	order.ID = result.InsertedID.(bson.ObjectID)
	writeJSON(w, http.StatusCreated, order)
}
