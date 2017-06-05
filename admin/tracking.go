package admin

import (
	"time"
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"bloodtales/controllers"
	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

func handleAdminTracking() {
	handleAdminTemplate("/admin/trackings", system.TokenAuthentication, ViewTrackings, "trackings.tmpl.html")
	handleAdminTemplate("/admin/trackings/view", system.TokenAuthentication, ViewTracking, "tracking.tmpl.html")
	handleAdminTemplate("/admin/trackings/delete", system.TokenAuthentication, DeleteTracking, "")

	// HACK
	system.App.HandleAPI("/admin/trackings/test", system.TokenAuthentication, TestTrackings)
}

// HACK
func TestTrackings(context *util.Context) {
	controllers.InsertTracking(context, "test", bson.M { "testArg": time.Now(), "testArg2": 30 }, 1)
}

func ViewTrackings(context *util.Context) {
	// paginate players query
	pagination, err := context.Paginate(context.DB.C(models.TrackingCollectionName).Find(nil).Sort("-t0"), DefaultPageSize)
	util.Must(err)

	// get resulting trackings
	var trackings []*models.Tracking
	util.Must(pagination.All(&trackings))

	// set template bindings
	context.Params.Set("trackings", trackings)
}

func ViewTracking(context *util.Context) {
	// parse parameters
	trackingId := context.Params.GetRequiredId("trackingId")

	tracking, err := models.GetTrackingById(context, trackingId)
	util.Must(err)
	
	// set template bindings
	context.Params.Set("tracking", tracking)
}

func DeleteTracking(context *util.Context) {
	// parse parameters
	trackingId := context.Params.GetRequiredId("trackingId")
	page := context.Params.GetInt("page", 1)

	tracking, err := models.GetTrackingById(context, trackingId)
	util.Must(err)

	tracking.Delete(context)

	context.Redirect(fmt.Sprintf("/admin/trackings?page=%d", page), 302)
}
