package util

import "go.mongodb.org/mongo-driver/bson/primitive"

// GenerateMongoID 生成 MongoDB 原生的 ObjectID 并转为字符串
func GenerateMongoID() string {
	// 生成类似 "65a1f2b3..." 的 24 位十六进制字符串
	return primitive.NewObjectID().Hex()
}
