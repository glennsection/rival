package util

import (
	"fmt"
	"time"
	"runtime/debug"
	
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const FaultCollectionName = "faults"

type Fault struct {
	ID             bson.ObjectId `bson:"_id,omitempty" json:"-"`
	UserID         bson.ObjectId `bson:"us,omitempty" json:"-"`
	CreatedTime    time.Time     `bson:"t0" json:"created"`
	ExpireTime     time.Time     `bson:"exp,omitempty" json:"expires"`
	Error          string        `bson:"err" json:"error"`
	Stack          string        `bson:"st,omitempty" json:"stack"`
}

func EnsureIndexFault(database *mgo.Database) {
	c := database.C(FaultCollectionName)

	// user index
	Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "us" },
		Background:   true,
		Sparse:       true,
	}))

	// expiration
	Must(c.EnsureIndex(mgo.Index {
		Key:          []string { "exp" },
		Background:   true,
		Sparse:       true,
		ExpireAfter:  1,
	}))
}

func GetFaultById(context *Context, id bson.ObjectId) (fault *Fault, err error) {
	err = context.DB.C(FaultCollectionName).Find(bson.M { "_id": id } ).One(&fault)
	return
}

func InsertErrorFault(context *Context, err interface{}) error {
	// get stack
	var stack string
	if errStack, ok := err.(*errorStack); ok {
		stack = string(errStack.stack)
	} else {
		stack = string(debug.Stack())
	}

	fault := &Fault {
		UserID: context.UserID,
		ExpireTime: time.Now().Add(time.Hour * time.Duration(168)), // 1 week
		Error: fmt.Sprintf("%v", err),
		Stack: stack,
	}

	return fault.Insert(context)
}

func (fault *Fault) Insert(context *Context) (err error) {
	fault.ID = bson.NewObjectId()
	fault.CreatedTime = time.Now()
	err = context.DB.C(FaultCollectionName).Insert(fault)
	return
}

func (fault *Fault) Delete(context *Context) (err error) {
	return context.DB.C(FaultCollectionName).Remove(bson.M { "_id": fault.ID })
}

func GetFaults(context *Context, userId bson.ObjectId) (faults *[]Fault, err error) {
	err = context.DB.C(FaultCollectionName).Find(bson.M { "us": userId } ).Sort("t0").All(&faults)
	return
}
