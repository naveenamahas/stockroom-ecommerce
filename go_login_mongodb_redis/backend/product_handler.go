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

const productCacheTTL = 5 * time.Minute

func productsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ctx, cancel, ok := sessionUserID(w, r)
	if !ok {
		return
	}
	defer cancel()
	if r.URL.Path != "/api/products" {
		productByIDHandler(w, r, userID, ctx)
		return
	}
	switch r.Method {
	case http.MethodGet:
		cacheKey := "products:" + userID.Hex()
		if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
			var items []Product
			if json.Unmarshal([]byte(cached), &items) == nil {
				writeJSON(w, http.StatusOK, items)
				return
			}
		}
		cursor, err := products.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
		if err != nil {
			errorJSON(w, 500, "could not load products")
			return
		}
		defer cursor.Close(ctx)
		items := []Product{}
		if err = cursor.All(ctx, &items); err != nil {
			errorJSON(w, 500, "could not read products")
			return
		}
		if encoded, err := json.Marshal(items); err == nil {
			_ = rdb.Set(ctx, cacheKey, encoded, productCacheTTL).Err()
		}
		writeJSON(w, 200, items)
	case http.MethodPost:
		var request ProductRequest
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			errorJSON(w, 400, "invalid JSON")
			return
		}
		if !validProduct(request) {
			errorJSON(w, 400, "name, positive price and quantity are required")
			return
		}
		now := time.Now()
		product := Product{UserID: userID, Name: strings.TrimSpace(request.Name), Description: strings.TrimSpace(request.Description), Price: request.Price, Quantity: request.Quantity, CreatedAt: now, UpdatedAt: now}
		result, err := products.InsertOne(ctx, product)
		if err != nil {
			errorJSON(w, 500, "could not create product")
			return
		}
		product.ID = result.InsertedID.(bson.ObjectID)
		invalidateProductCache(ctx, userID)
		writeJSON(w, 201, product)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func productByIDHandler(w http.ResponseWriter, r *http.Request, userID bson.ObjectID, ctx context.Context) {
	productID, err := bson.ObjectIDFromHex(strings.TrimPrefix(r.URL.Path, "/api/products/"))
	if err != nil {
		errorJSON(w, 400, "invalid product id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var request ProductRequest
		if json.NewDecoder(r.Body).Decode(&request) != nil {
			errorJSON(w, 400, "invalid JSON")
			return
		}
		if !validProduct(request) {
			errorJSON(w, 400, "name, positive price and quantity are required")
			return
		}
		update := bson.M{"$set": bson.M{"name": strings.TrimSpace(request.Name), "description": strings.TrimSpace(request.Description), "price": request.Price, "quantity": request.Quantity, "updatedAt": time.Now()}}
		var product Product
		err = products.FindOneAndUpdate(ctx, bson.M{"_id": productID, "userId": userID}, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&product)
		if err == mongo.ErrNoDocuments {
			errorJSON(w, 404, "product not found")
			return
		}
		if err != nil {
			errorJSON(w, 500, "could not update product")
			return
		}
		invalidateProductCache(ctx, userID)
		writeJSON(w, 200, product)
	case http.MethodDelete:
		result, err := products.DeleteOne(ctx, bson.M{"_id": productID, "userId": userID})
		if err != nil {
			errorJSON(w, 500, "could not delete product")
			return
		}
		if result.DeletedCount == 0 {
			errorJSON(w, 404, "product not found")
			return
		}
		invalidateProductCache(ctx, userID)
		writeJSON(w, 200, map[string]string{"message": "product deleted"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func validProduct(product ProductRequest) bool {
	return strings.TrimSpace(product.Name) != "" && product.Price >= 0 && product.Quantity >= 0
}

func invalidateProductCache(ctx context.Context, userID bson.ObjectID) {
	_ = rdb.Del(ctx, "products:"+userID.Hex()).Err()
}
